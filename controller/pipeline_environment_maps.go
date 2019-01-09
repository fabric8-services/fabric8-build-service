package controller

import (
	"context"
	"github.com/fabric8-services/fabric8-build/app"
	"github.com/fabric8-services/fabric8-build/application"
	"github.com/fabric8-services/fabric8-build/application/env"
	"github.com/fabric8-services/fabric8-build/build"
	"github.com/fabric8-services/fabric8-common/errors"
	"github.com/fabric8-services/fabric8-common/httpsupport"
	"github.com/fabric8-services/fabric8-common/token"
	"github.com/goadesign/goa"
	guuid "github.com/goadesign/goa/uuid"
	errs "github.com/pkg/errors"
	"github.com/prometheus/common/log"
)

// PipelineEnvironmentMapsController implements the PipelineEnvironmentMaps resource.
type PipelineEnvironmentMapsController struct {
	*goa.Controller
	db         application.DB
	svcFactory application.ServiceFactory
}

// NewPipelineEnvironmentMapsController creates a PipelineEnvironmentMaps controller.
func NewPipelineEnvironmentMapsController(service *goa.Service, db application.DB, svcFactory application.ServiceFactory) *PipelineEnvironmentMapsController {
	return &PipelineEnvironmentMapsController{
		Controller: service.NewController("PipelineEnvironmentControllerMap"),
		db:         db,
		svcFactory: svcFactory,
	}
}

// Create runs the create action.
func (c *PipelineEnvironmentMapsController) Create(ctx *app.CreatePipelineEnvironmentMapsContext) error {
	tokenMgr, err := token.ReadManagerFromContext(ctx)
	if err != nil {
		return app.JSONErrorResponse(ctx, err)
	}
	_, err = tokenMgr.Locate(ctx)
	if err != nil {
		return app.JSONErrorResponse(ctx, errors.NewUnauthorizedError(err.Error()))
	}

	err = validateCreatePipelineEnvironmentMap(ctx)
	if err != nil {
		return app.JSONErrorResponse(ctx, err)
	}

	reqPpl := ctx.Payload.Data
	spaceID := ctx.SpaceID
	err = c.checkSpaceExist(ctx, spaceID.String())
	if err != nil {
		return app.JSONErrorResponse(ctx, err)
	}

	newEnvs, err := c.checkEnvironmentExistAndConvert(ctx, spaceID.String(), reqPpl.Environments)
	if err != nil {
		return app.JSONErrorResponse(ctx, errors.NewNotFoundError("environment", err.Error()))
	}

	var ppl *build.PipelineEnvMap
	err = application.Transactional(c.db, func(appl application.Application) error {
		newPipeline := build.PipelineEnvMap{
			Name:         &reqPpl.Name,
			SpaceID:      &spaceID,
			Environments: newEnvs,
		}

		ppl, err = appl.PipelineEnvMap().Create(ctx, &newPipeline)
		if err != nil {
			log.Error(ctx, map[string]interface{}{"err": err},
				"failed to create pipelineenvmap: %s", newPipeline.Name)
			return errs.Wrapf(err, "failed to create pipelineenvmap: %s", *newPipeline.Name)
		}

		return nil
	})

	if err != nil {
		return app.JSONErrorResponse(ctx, err)
	}

	newEnvAttributes := []*app.EnvironmentAttributes{}
	for _, pipeline := range ppl.Environments {
		newEnvAttributes = append(newEnvAttributes, &app.EnvironmentAttributes{
			EnvUUID: pipeline.EnvironmentID,
		})
	}

	res := &app.PipelineEnvironmentMapSingle{
		Data: &app.PipelineEnvironmentMaps{
			ID:           &ppl.ID,
			Name:         *ppl.Name,
			Environments: newEnvAttributes,
			SpaceID:      ppl.SpaceID,
		},
	}

	ctx.ResponseData.Header().Set(
		"Location",
		httpsupport.AbsoluteURL(&goa.RequestData{Request: ctx.Request}, app.PipelineEnvironmentMapsHref(res.Data.ID), nil),
	)
	return ctx.Created(res)
}

// List runs the list action.
func (c *PipelineEnvironmentMapsController) List(ctx *app.ListPipelineEnvironmentMapsContext) error {
	spaceID := ctx.SpaceID
	err := c.checkSpaceExist(ctx, spaceID.String())
	if err != nil {
		return app.JSONErrorResponse(ctx, err)
	}

	pplenvmaps, err := c.db.PipelineEnvMap().List(ctx, spaceID)
	if err != nil {
		return app.JSONErrorResponse(ctx, err)
	}

	newPipelineEnvMapList := []*app.PipelineEnvironmentMaps{}
	for _, pipEnvMap := range pplenvmaps {
		newPipelineEnvMapList = append(newPipelineEnvMapList, convertToPipelineEnvironmentMapStruct(pipEnvMap))
	}

	res := &app.PipelineEnvironmentMapsList{
		Data: newPipelineEnvMapList,
	}
	return ctx.OK(res)
}

// Show runs the load action.
func (c *PipelineEnvironmentMapsController) Show(ctx *app.ShowPipelineEnvironmentMapsContext) error {
	envID := ctx.ID
	pipenvmap, err := c.db.PipelineEnvMap().Load(ctx, envID)
	if err != nil {
		return app.JSONErrorResponse(ctx, err)
	}

	data := convertToPipelineEnvironmentMapStruct(pipenvmap)
	res := &app.PipelineEnvironmentMapSingle{
		Data: data,
	}
	return ctx.OK(res)
}

// Update runs the save action.
func (c *PipelineEnvironmentMapsController) Update(ctx *app.UpdatePipelineEnvironmentMapsContext) error {
	tokenMgr, err := token.ReadManagerFromContext(ctx)
	if err != nil {
		return app.JSONErrorResponse(ctx, err)
	}
	_, err = tokenMgr.Locate(ctx)
	if err != nil {
		return app.JSONErrorResponse(ctx, errors.NewUnauthorizedError(err.Error()))
	}

	err = validateUpdatePipelineEnvironmentMap(ctx)
	if err != nil {
		return app.JSONErrorResponse(ctx, err)
	}

	reqPpl := ctx.Payload.Data
	spaceID := reqPpl.SpaceID
	err = c.checkSpaceExist(ctx, spaceID.String())
	if err != nil {
		return app.JSONErrorResponse(ctx, err)
	}

	newEnvs, err := c.checkEnvironmentExistAndConvert(ctx, spaceID.String(), reqPpl.Environments)
	if err != nil {
		return app.JSONErrorResponse(ctx, errors.NewNotFoundError("environment", err.Error()))
	}

	var ppl *build.PipelineEnvMap
	err = application.Transactional(c.db, func(appl application.Application) error {
		ppl, err = c.db.PipelineEnvMap().Load(ctx, ctx.ID)
		if err != nil {
			return app.JSONErrorResponse(ctx, err)
		}

		ppl.Name = &ctx.Payload.Data.Name
		ppl.Environments = newEnvs
		ppl, err = appl.PipelineEnvMap().Save(ctx, ppl)
		return err
	})
	if err != nil {
		return app.JSONErrorResponse(ctx, err)
	}

	data := convertToPipelineEnvironmentMapStruct(ppl)
	res := &app.PipelineEnvironmentMapSingle{
		Data: data,
	}
	return ctx.OK(res)
}

// This will check whether the given space exist or not
func (c *PipelineEnvironmentMapsController) checkSpaceExist(ctx context.Context, spaceID string) error {
	// TODO(chmouel): Make sure we have the rights for that space
	// TODO(chmouel): Better error reporting when NOTFound
	_, err := c.svcFactory.WITService().GetSpace(ctx, spaceID)
	if err != nil {
		return errs.Wrapf(err, "failed to get space id: %s from wit", spaceID)
	}
	return nil
}

// This will check whether the env's exit and then convert to build.Environment List
func (c *PipelineEnvironmentMapsController) checkEnvironmentExistAndConvert(ctx context.Context, spaceID string, envs []*app.EnvironmentAttributes) ([]build.PipelineEnvironment, error) {
	envList, err := c.svcFactory.ENVService().GetEnvList(ctx, spaceID)
	if err != nil {
		return nil, errs.Wrapf(err, "failed to get env list for space id: %s from env service", spaceID)
	}
	envUUIDList := convertToEnvUIDList(envList)
	var environments []build.PipelineEnvironment
	for _, env := range envs {
		envID, _ := guuid.FromString(env.EnvUUID.String())
		envName := envUUIDList[envID]
		if envName == "" {
			return nil, errors.NewNotFoundError("environment", env.EnvUUID.String())
		}
		environments = append(environments, build.PipelineEnvironment{
			EnvironmentID: env.EnvUUID,
		})
	}
	return environments, nil
}

// This will convert the list of env into map[envId]envName
// this will help to check whether env exist or not
func convertToEnvUIDList(envList []env.Environment) map[guuid.UUID]string {
	var envMap = make(map[guuid.UUID]string)
	for _, env := range envList {
		envMap[env.ID] = env.Name
	}
	return envMap
}

// this will convert the pipeline struct from database to pipeline-environment struct
func convertToPipelineEnvironmentMapStruct(ppl *build.PipelineEnvMap) *app.PipelineEnvironmentMaps {
	newEnvAttributes := []*app.EnvironmentAttributes{}
	for _, pipelineEnv := range ppl.Environments {
		newEnvAttributes = append(newEnvAttributes, &app.EnvironmentAttributes{
			EnvUUID: pipelineEnv.EnvironmentID,
		})
	}

	pe := &app.PipelineEnvironmentMaps{
		ID:           &ppl.ID,
		Name:         *ppl.Name,
		Environments: newEnvAttributes,
		SpaceID:      ppl.SpaceID,
	}
	return pe
}

func validateCreatePipelineEnvironmentMap(ctx *app.CreatePipelineEnvironmentMapsContext) error {
	if ctx.Payload.Data == nil {
		return errors.NewBadParameterError("data", nil).Expected("not nil")
	}
	if ctx.Payload.Data.SpaceID == nil {
		return errors.NewBadParameterError("data.spaceId", nil).Expected("not nil")
	}
	if ctx.Payload.Data.Name == "" {
		return errors.NewBadParameterError("data.name", nil).Expected("not nil")
	}
	if ctx.Payload.Data.Environments == nil || len(ctx.Payload.Data.Environments) == 0 {
		return errors.NewBadParameterError("data.environments", nil).Expected("not nil")
	}
	return nil
}

func validateUpdatePipelineEnvironmentMap(ctx *app.UpdatePipelineEnvironmentMapsContext) error {
	if ctx.Payload.Data == nil {
		return errors.NewBadParameterError("data", nil).Expected("not nil")
	}
	if ctx.Payload.Data.SpaceID == nil {
		return errors.NewBadParameterError("data.spaceId", nil).Expected("not nil")
	}
	if ctx.Payload.Data.Name == "" {
		return errors.NewBadParameterError("data.name", nil).Expected("not nil")
	}
	if ctx.Payload.Data.Environments == nil || len(ctx.Payload.Data.Environments) == 0 {
		return errors.NewBadParameterError("data.environments", nil).Expected("not nil")
	}
	return nil
}
