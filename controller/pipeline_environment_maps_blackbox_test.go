package controller_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/fabric8-services/fabric8-build/app"
	"github.com/fabric8-services/fabric8-build/app/test"
	"github.com/fabric8-services/fabric8-build/application"
	"github.com/fabric8-services/fabric8-build/application/env/envservice"
	"github.com/fabric8-services/fabric8-build/application/wit/witservice"
	"github.com/fabric8-services/fabric8-build/configuration"
	"github.com/fabric8-services/fabric8-build/controller"
	"github.com/fabric8-services/fabric8-build/gormapp"
	testauth "github.com/fabric8-services/fabric8-common/test/auth"
	testsuite "github.com/fabric8-services/fabric8-common/test/suite"
	"github.com/goadesign/goa"
	guuid "github.com/goadesign/goa/uuid"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/h2non/gock.v1"
)

type PipelineEnvironmentMapsControllerSuite struct {
	testsuite.DBTestSuite
	db *gormapp.GormDB

	svc  *goa.Service // secure
	svc2 *goa.Service // unsecure
	ctx  context.Context
	ctx2 context.Context

	ctrl  *controller.PipelineEnvironmentMapsController
	ctrl2 *controller.PipelineEnvironmentMapsController

	svcFactory application.ServiceFactory
}

func TestPipelineEnvironmentMapsController(t *testing.T) {
	config, err := configuration.New("")
	require.NoError(t, err)
	suite.Run(t, &PipelineEnvironmentMapsControllerSuite{DBTestSuite: testsuite.NewDBTestSuite(config)})
}

func (s *PipelineEnvironmentMapsControllerSuite) SetupSuite() {
	s.DBTestSuite.SetupSuite()
	config, _ := configuration.New("")

	s.db = gormapp.NewGormDB(s.DB)

	svc := testauth.UnsecuredService("ppl-test1")
	s.svc = svc

	svc2, err := testauth.ServiceAsUser("ppl-test2", testauth.NewIdentity())
	require.NoError(s.T(), err)
	s.svc2 = svc2

	s.ctx = s.svc.Context
	s.ctx2 = s.svc2.Context

	s.svcFactory = application.NewServiceFactory(config)

	s.ctrl = controller.NewPipelineEnvironmentMapsController(s.svc, s.db, s.svcFactory)
	s.ctrl2 = controller.NewPipelineEnvironmentMapsController(s.svc2, s.db, s.svcFactory)

	os.Setenv("F8_WIT_URL", "http://witservice")
	os.Setenv("F8_ENV_URL", "http://envservice")
	// gock.Observe(gock.DumpRequest)

	defer gock.OffAll()
}

func (s *PipelineEnvironmentMapsControllerSuite) createSpaceJson(spaceName string, spaceID uuid.UUID) string {
	// TODO: Test ownership
	identityID := guuid.NewV4()
	desc := "Description of " + spaceName
	version := 0
	spaceTime := time.Now()
	_spaceID, _ := guuid.FromString(spaceID.String())

	wt := witservice.SpaceSingle{
		Data: &witservice.Space{
			ID: &_spaceID,
			Attributes: &witservice.SpaceAttributes{
				CreatedAt:   &spaceTime,
				Description: &desc,
				Name:        &spaceName,
				UpdatedAt:   &spaceTime,
				Version:     &version,
			},
			Links: &witservice.GenericLinksForSpace{},
			Type:  "spaces",
			Relationships: &witservice.SpaceRelationships{
				OwnedBy: &witservice.SpaceOwnedBy{
					Data: &witservice.IdentityRelationData{
						ID:   &identityID,
						Type: "identities",
					},
				},
			},
		},
	}

	b, _ := json.Marshal(wt)
	return string(b)
}

func (s *PipelineEnvironmentMapsControllerSuite) createEnvListJson(envID1 uuid.UUID, envID2 uuid.UUID) string {
	_envID1, _ := guuid.FromString(envID1.String())
	_envID2, _ := guuid.FromString(envID2.String())
	_envName1 := "env1"
	_envName2 := "env2"

	env := envservice.EnvironmentsList{
		Data: []*envservice.Environment{
			{
				ID: &_envID1,
				Attributes: &envservice.EnvironmentAttributes{
					Name: &_envName1,
				},
				Links: &envservice.GenericLinks{},
				Type:  "environments",
			},
			{
				ID: &_envID2,
				Attributes: &envservice.EnvironmentAttributes{
					Name: &_envName2,
				},
				Links: &envservice.GenericLinks{},
				Type:  "environments",
			},
		},
		Links: &envservice.PagingLinks{},
		Meta:  &envservice.EnvironmentListMeta{},
	}
	b, _ := json.Marshal(env)
	return string(b)
}

func (s *PipelineEnvironmentMapsControllerSuite) createGockONSpace(spaceID uuid.UUID, spaceName string) {
	gock.New("http://witservice").
		Get("/api/spaces/" + spaceID.String()).
		Reply(200).
		JSON(s.createSpaceJson(spaceName, spaceID))
}

func (s *PipelineEnvironmentMapsControllerSuite) createGockONEnvList(spaceID uuid.UUID, envID1 uuid.UUID, envID2 uuid.UUID) {
	gock.New("http://envservice").
		Get("/api/spaces/" + spaceID.String() + "/environments").
		Reply(200).
		JSON(s.createEnvListJson(envID1, envID2))
}

// createPipelineEnvironmentCtrlNoErroring we do this one manually cause the one from
// goatest one exit on errro without being able to catch
func (s *PipelineEnvironmentMapsControllerSuite) createPipelineEnvironmentCtrlNoErroring(spaceID uuid.UUID) (*app.CreatePipelineEnvironmentMapsContext, *httptest.ResponseRecorder) {
	rw := httptest.NewRecorder()
	u := &url.URL{
		Path: fmt.Sprintf("/api/spaces/%v/pipeline-environment-maps", spaceID),
	}
	req, _err := http.NewRequest("POST", u.String(), nil)
	if _err != nil {
		panic("invalid test " + _err.Error()) // bug
	}
	prms := url.Values{}
	prms["spaceID"] = []string{fmt.Sprintf("%v", spaceID)}
	goaCtx := goa.NewContext(goa.WithAction(s.ctx2, "PipelineEnvironmentsTest"), rw, req, prms)
	createCtx, __err := app.NewCreatePipelineEnvironmentMapsContext(goaCtx, req, s.svc2)
	if __err != nil {
		panic("invalid test data " + __err.Error()) // bug
	}
	return createCtx, rw
}

func (s *PipelineEnvironmentMapsControllerSuite) TestCreate() {
	defer s.T().Run("ok", func(t *testing.T) {
		space1ID := uuid.NewV4()
		env1ID := uuid.NewV4()
		env2ID := uuid.NewV4()
		s.createGockONSpace(space1ID, "space1")
		s.createGockONEnvList(space1ID, env1ID, env2ID)
		payload := newPipelineEnvironmentMapPayload("osio-stage-create", space1ID, env1ID)
		_, newEnv := test.CreatePipelineEnvironmentMapsCreated(t, s.ctx2, s.svc2, s.ctrl2, space1ID, payload)
		assert.NotNil(t, newEnv)
		assert.NotNil(t, newEnv.Data.ID)
		assert.NotNil(t, newEnv.Data.Environments[0].EnvUUID)

		// Same pipeline_name but different spaceID is OK
		space2ID := uuid.NewV4()
		s.createGockONSpace(space2ID, "space2")
		s.createGockONEnvList(space2ID, env1ID, env2ID)
		payload = newPipelineEnvironmentMapPayload("osio-stage-create", space2ID, env1ID)
		_, newEnv = test.CreatePipelineEnvironmentMapsCreated(t, s.ctx2, s.svc2, s.ctrl2, space2ID, payload)
		assert.NotNil(t, newEnv)
		assert.NotNil(t, newEnv.Data.ID)
		assert.NotNil(t, newEnv.Data.Environments[0].EnvUUID)
	})

	s.T().Run("fail", func(t *testing.T) {
		space1ID := uuid.NewV4()
		env1ID := uuid.NewV4()
		env2ID := uuid.NewV4()

		s.createGockONSpace(space1ID, "space1")
		s.createGockONEnvList(space1ID, env1ID, env2ID)
		payload := newPipelineEnvironmentMapPayload("osio-stage-create-conflict", space1ID, env1ID)
		_, newEnv := test.CreatePipelineEnvironmentMapsCreated(t, s.ctx2, s.svc2, s.ctrl2, space1ID, payload)
		assert.NotNil(t, newEnv)

		s.createGockONSpace(space1ID, "space1")
		s.createGockONEnvList(space1ID, env1ID, env2ID)
		response, err := test.CreatePipelineEnvironmentMapsConflict(t, s.ctx2, s.svc2, s.ctrl2, space1ID, payload)
		require.NotNil(t, response.Header().Get("Location"))
		assert.Regexp(s.T(), ".*data_conflict_error.*", err.Errors)

		emptyPayload := &app.CreatePipelineEnvironmentMapsPayload{}
		createCtxerr, rw := s.createPipelineEnvironmentCtrlNoErroring(space1ID)
		createCtxerr.Payload = emptyPayload
		jerr := s.ctrl2.Create(createCtxerr)
		require.Nil(t, jerr)
		require.Equal(t, 400, rw.Code)

		failSpaceID := uuid.NewV4()
		gock.New("http://witservice").
			Get("/api/spaces/" + failSpaceID.String()).
			Reply(404)
		payload = newPipelineEnvironmentMapPayload("space-not-found", failSpaceID, env1ID)
		response, err = test.CreatePipelineEnvironmentMapsNotFound(t, s.ctx2, s.svc2, s.ctrl2, failSpaceID, payload)
		require.NotNil(t, response.Header().Get("Location"))
		assert.Regexp(s.T(), ".*not_found.*", err.Errors)

		failSpaceID = uuid.NewV4()
		gock.New("http://witservice").
			Get("/api/spaces/" + failSpaceID.String()).
			Reply(422)
		payload = newPipelineEnvironmentMapPayload("space-unkown-error", failSpaceID, env1ID)
		response, err = test.CreatePipelineEnvironmentMapsInternalServerError(t, s.ctx2, s.svc2, s.ctrl2, failSpaceID, payload)
		require.NotNil(t, response.Header().Get("Location"))
		assert.Regexp(s.T(), ".*unknown_error.*", err.Errors)

		failEnvID := uuid.NewV4()
		s.createGockONSpace(space1ID, "space1")
		s.createGockONEnvList(space1ID, env1ID, env2ID)
		payload = newPipelineEnvironmentMapPayload("env-not-found", space1ID, failEnvID)
		response, err = test.CreatePipelineEnvironmentMapsNotFound(t, s.ctx2, s.svc2, s.ctrl2, space1ID, payload)
		require.NotNil(t, response.Header().Get("Location"))
		assert.Regexp(s.T(), ".*not_found.*", err.Errors)
	})

	s.T().Run("unauthorized", func(t *testing.T) {
		space1ID := uuid.NewV4()
		env1ID := uuid.NewV4()
		env2ID := uuid.NewV4()
		s.createGockONSpace(space1ID, "space1")
		s.createGockONEnvList(space1ID, env1ID, env2ID)
		payload := newPipelineEnvironmentMapPayload("osio-stage", space1ID, env2ID)
		_, err := test.CreatePipelineEnvironmentMapsUnauthorized(t, s.ctx, s.svc, s.ctrl, space1ID, payload)
		assert.NotNil(t, err)
	})
}

func (s *PipelineEnvironmentMapsControllerSuite) TestShow() {
	s.T().Run("ok", func(t *testing.T) {
		spaceID := uuid.NewV4()
		env1ID := uuid.NewV4()
		env2ID := uuid.NewV4()
		s.createGockONSpace(spaceID, "space1")
		s.createGockONEnvList(spaceID, env1ID, env2ID)
		payload := newPipelineEnvironmentMapPayload("osio-stage-show", spaceID, env1ID)
		_, newEnv := test.CreatePipelineEnvironmentMapsCreated(t, s.ctx2, s.svc2, s.ctrl2, spaceID, payload)
		require.NotNil(t, newEnv)

		_, env := test.ShowPipelineEnvironmentMapsOK(t, s.ctx2, s.svc2, s.ctrl2, *newEnv.Data.ID)
		assert.NotNil(t, env)
		assert.Equal(t, newEnv.Data.ID, env.Data.ID)
	})

	s.T().Run("not_found", func(t *testing.T) {
		spaceID := uuid.NewV4()
		env1ID := uuid.NewV4()
		env2ID := uuid.NewV4()
		s.createGockONSpace(spaceID, "space1")
		s.createGockONEnvList(spaceID, env1ID, env2ID)
		payload := newPipelineEnvironmentMapPayload("osio-stage-show", spaceID, env1ID)
		_, newEnv := test.CreatePipelineEnvironmentMapsCreated(t, s.ctx2, s.svc2, s.ctrl2, spaceID, payload)
		require.NotNil(t, newEnv)

		_, err := test.ShowPipelineEnvironmentMapsNotFound(t, s.ctx2, s.svc2, s.ctrl2, uuid.NewV4())
		assert.NotNil(t, err)
	})
}

func (s *PipelineEnvironmentMapsControllerSuite) TestList() {
	s.T().Run("ok", func(t *testing.T) {
		spaceID := uuid.NewV4()
		env1ID := uuid.NewV4()
		env2ID := uuid.NewV4()
		s.createGockONSpace(spaceID, "space1")
		s.createGockONEnvList(spaceID, env1ID, env2ID)
		payload := newPipelineEnvironmentMapPayload("osio-stage-show", spaceID, env1ID)
		_, newEnv := test.CreatePipelineEnvironmentMapsCreated(t, s.ctx2, s.svc2, s.ctrl2, spaceID, payload)
		require.NotNil(t, newEnv)
		s.createGockONSpace(spaceID, "space1")
		s.createGockONEnvList(spaceID, env1ID, env2ID)
		payload2 := newPipelineEnvironmentMapPayload("osio-stage-show2", spaceID, env2ID)
		_, newEnv2 := test.CreatePipelineEnvironmentMapsCreated(t, s.ctx2, s.svc2, s.ctrl2, spaceID, payload2)
		require.NotNil(t, newEnv2)

		s.createGockONSpace(spaceID, "space1")
		_, env := test.ListPipelineEnvironmentMapsOK(t, s.ctx2, s.svc2, s.ctrl2, spaceID)
		assert.NotNil(t, env)
		assert.Equal(t, 2, len(env.Data))
	})

	s.T().Run("space_not_found", func(t *testing.T) {
		spaceID := uuid.NewV4()
		_, err := test.ListPipelineEnvironmentMapsInternalServerError(t, s.ctx2, s.svc2, s.ctrl2, spaceID)
		assert.NotNil(t, err)
	})
}

func (s *PipelineEnvironmentMapsControllerSuite) TestUpdate() {
	s.T().Run("ok", func(t *testing.T) {
		spaceID := uuid.NewV4()
		env1ID := uuid.NewV4()
		env2ID := uuid.NewV4()
		s.createGockONSpace(spaceID, "space1")
		s.createGockONEnvList(spaceID, env1ID, env2ID)
		payload := newPipelineEnvironmentMapPayload("osio-stage-update", spaceID, env1ID)
		_, newEnv := test.CreatePipelineEnvironmentMapsCreated(t, s.ctx2, s.svc2, s.ctrl2, spaceID, payload)
		require.NotNil(t, newEnv)
		s.createGockONSpace(spaceID, "space1")
		s.createGockONEnvList(spaceID, env1ID, env2ID)
		payload2 := updatePipelineEnvironmentMapPayload(payload, env2ID)
		_, newEnv2 := test.UpdatePipelineEnvironmentMapsOK(t, s.ctx2, s.svc2, s.ctrl2, *newEnv.Data.ID, payload2)
		require.NotNil(t, newEnv2)
		assert.Equal(t, env2ID, *(newEnv2.Data.Environments[0].EnvUUID))
	})

	s.T().Run("unauthorized", func(t *testing.T) {
		space1ID := uuid.NewV4()
		env1ID := uuid.NewV4()
		env2ID := uuid.NewV4()
		s.createGockONSpace(space1ID, "space1")
		s.createGockONEnvList(space1ID, env1ID, env2ID)
		payload := newPipelineEnvironmentMapPayload("osio-stage", space1ID, env2ID)
		_, env := test.CreatePipelineEnvironmentMapsCreated(t, s.ctx2, s.svc2, s.ctrl2, space1ID, payload)
		assert.NotNil(t, env)
		s.createGockONSpace(space1ID, "space1")
		s.createGockONEnvList(space1ID, env1ID, env2ID)
		uPayload := updatePipelineEnvironmentMapPayload(payload, env1ID)
		_, err := test.UpdatePipelineEnvironmentMapsUnauthorized(t, s.ctx, s.svc, s.ctrl2, *env.Data.ID, uPayload)
		assert.NotNil(t, err)
	})
}

func newPipelineEnvironmentMapPayload(name string, spaceID uuid.UUID, envUUID uuid.UUID) *app.CreatePipelineEnvironmentMapsPayload {
	payload := &app.CreatePipelineEnvironmentMapsPayload{
		Data: &app.PipelineEnvironmentMaps{
			Name:    name,
			SpaceID: &spaceID,
			Environments: []*app.EnvironmentAttributes{
				{EnvUUID: &envUUID},
			},
		},
	}
	return payload
}

func updatePipelineEnvironmentMapPayload(pEnv *app.CreatePipelineEnvironmentMapsPayload, envUUID uuid.UUID) *app.UpdatePipelineEnvironmentMapsPayload {
	payload := &app.UpdatePipelineEnvironmentMapsPayload{
		Data: &app.PipelineEnvironmentMaps{
			Name:    pEnv.Data.Name,
			SpaceID: pEnv.Data.SpaceID,
			ID:      pEnv.Data.ID,
			Environments: []*app.EnvironmentAttributes{
				{EnvUUID: &envUUID},
			},
		},
	}
	return payload
}
