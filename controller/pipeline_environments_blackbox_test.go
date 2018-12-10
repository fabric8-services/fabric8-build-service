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

type PipelineEnvironmentControllerSuite struct {
	testsuite.DBTestSuite
	db *gormapp.GormDB

	svc  *goa.Service // secure
	svc2 *goa.Service // unsecure
	ctx  context.Context
	ctx2 context.Context

	ctrl  *controller.PipelineEnvironmentController
	ctrl2 *controller.PipelineEnvironmentController

	svcFactory application.ServiceFactory
}

func TestPipelineEnvironmentController(t *testing.T) {
	config, err := configuration.New("")
	require.NoError(t, err)
	suite.Run(t, &PipelineEnvironmentControllerSuite{DBTestSuite: testsuite.NewDBTestSuite(config)})
}

func (s *PipelineEnvironmentControllerSuite) SetupSuite() {
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

	s.ctrl = controller.NewPipelineEnvironmentController(s.svc, s.db, s.svcFactory)
	s.ctrl2 = controller.NewPipelineEnvironmentController(s.svc2, s.db, s.svcFactory)

	os.Setenv("F8_WIT_URL", "http://witservice")
	// gock.Observe(gock.DumpRequest)

	defer gock.OffAll()
}

func (s *PipelineEnvironmentControllerSuite) createSpaceJson(spaceName string, spaceID uuid.UUID) string {
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

func (s *PipelineEnvironmentControllerSuite) createGockONSpace(spaceID uuid.UUID, spaceName string) {
	gock.New("http://witservice").
		Get("/api/spaces/" + spaceID.String()).
		Reply(200).
		JSON(s.createSpaceJson(spaceName, spaceID))
}

// createPipelineEnvironmentCtrlNoErroring we do this one manually cause the one from
// goatest one exit on errro without being able to catch
func (s *PipelineEnvironmentControllerSuite) createPipelineEnvironmentCtrlNoErroring(spaceID uuid.UUID) (*app.CreatePipelineEnvironmentsContext, *httptest.ResponseRecorder) {
	rw := httptest.NewRecorder()
	u := &url.URL{
		Path: fmt.Sprintf("/api/spaces/%v/pipeline-environments", spaceID),
	}
	req, _err := http.NewRequest("POST", u.String(), nil)
	if _err != nil {
		panic("invalid test " + _err.Error()) // bug
	}
	prms := url.Values{}
	prms["spaceID"] = []string{fmt.Sprintf("%v", spaceID)}
	goaCtx := goa.NewContext(goa.WithAction(s.ctx2, "PipelineEnvironmentsTest"), rw, req, prms)
	createCtx, __err := app.NewCreatePipelineEnvironmentsContext(goaCtx, req, s.svc2)
	if __err != nil {
		panic("invalid test data " + __err.Error()) // bug
	}
	return createCtx, rw
}

func (s *PipelineEnvironmentControllerSuite) TestCreate() {
	defer s.T().Run("ok", func(t *testing.T) {
		space1ID := uuid.NewV4()
		s.createGockONSpace(space1ID, "space1")
		payload := newPipelineEnvironmentPayload("osio-stage-create", uuid.NewV4())
		_, newEnv := test.CreatePipelineEnvironmentsCreated(t, s.ctx2, s.svc2, s.ctrl2, space1ID, payload)
		assert.NotNil(t, newEnv)
		assert.NotNil(t, newEnv.Data.ID)
		assert.NotNil(t, newEnv.Data.Environments[0].EnvUUID)

		// Same pipeline_name but different spaceID is OK
		space2ID := uuid.NewV4()
		s.createGockONSpace(space2ID, "space2")
		payload = newPipelineEnvironmentPayload("osio-stage-create", uuid.NewV4())
		_, newEnv = test.CreatePipelineEnvironmentsCreated(t, s.ctx2, s.svc2, s.ctrl2, space2ID, payload)
		assert.NotNil(t, newEnv)
		assert.NotNil(t, newEnv.Data.ID)
		assert.NotNil(t, newEnv.Data.Environments[0].EnvUUID)
	})

	s.T().Run("fail", func(t *testing.T) {
		space1ID := uuid.NewV4()

		s.createGockONSpace(space1ID, "space1")
		payload := newPipelineEnvironmentPayload("osio-stage-create-conflict", uuid.NewV4())
		_, newEnv := test.CreatePipelineEnvironmentsCreated(t, s.ctx2, s.svc2, s.ctrl2, space1ID, payload)
		assert.NotNil(t, newEnv)

		s.createGockONSpace(space1ID, "space1")
		response, err := test.CreatePipelineEnvironmentsConflict(t, s.ctx2, s.svc2, s.ctrl2, space1ID, payload)
		require.NotNil(t, response.Header().Get("Location"))
		assert.Regexp(s.T(), ".*data_conflict_error.*", err.Errors)

		emptyPayload := &app.CreatePipelineEnvironmentsPayload{}
		createCtxerr, rw := s.createPipelineEnvironmentCtrlNoErroring(space1ID)
		createCtxerr.Payload = emptyPayload
		jerr := s.ctrl2.Create(createCtxerr)
		require.Nil(t, jerr)
		require.Equal(t, 400, rw.Code)

		failSpaceID := uuid.NewV4()
		gock.New("http://witservice").
			Get("/api/spaces/" + failSpaceID.String()).
			Reply(404)
		payload = newPipelineEnvironmentPayload("space-not-found", uuid.NewV4())
		response, err = test.CreatePipelineEnvironmentsNotFound(t, s.ctx2, s.svc2, s.ctrl2, failSpaceID, payload)
		require.NotNil(t, response.Header().Get("Location"))
		assert.Regexp(s.T(), ".*not_found.*", err.Errors)

		failSpaceID = uuid.NewV4()
		gock.New("http://witservice").
			Get("/api/spaces/" + failSpaceID.String()).
			Reply(422)
		payload = newPipelineEnvironmentPayload("space-unkown-error", uuid.NewV4())
		response, err = test.CreatePipelineEnvironmentsInternalServerError(t, s.ctx2, s.svc2, s.ctrl2, failSpaceID, payload)
		require.NotNil(t, response.Header().Get("Location"))
		assert.Regexp(s.T(), ".*unknown_error.*", err.Errors)

	})

	s.T().Run("unauthorized", func(t *testing.T) {
		space1ID := uuid.NewV4()
		s.createGockONSpace(space1ID, "space1")

		payload := newPipelineEnvironmentPayload("osio-stage", uuid.NewV4())
		_, err := test.CreatePipelineEnvironmentsUnauthorized(t, s.ctx, s.svc, s.ctrl, space1ID, payload)
		assert.NotNil(t, err)
	})
}

func (s *PipelineEnvironmentControllerSuite) TestShow() {
	s.T().Run("ok", func(t *testing.T) {
		spaceID := uuid.NewV4()
		s.createGockONSpace(spaceID, "space1")
		payload := newPipelineEnvironmentPayload("osio-stage-show", uuid.NewV4())
		_, newEnv := test.CreatePipelineEnvironmentsCreated(t, s.ctx2, s.svc2, s.ctrl2, spaceID, payload)
		require.NotNil(t, newEnv)

		_, env := test.ShowPipelineEnvironmentsOK(t, s.ctx2, s.svc2, s.ctrl2, spaceID)
		assert.NotNil(t, env)
		assert.Equal(t, newEnv.Data.ID, env.Data.ID)
	})

	s.T().Run("not_found", func(t *testing.T) {
		envID := uuid.NewV4()
		_, err := test.ShowPipelineEnvironmentsNotFound(t, s.ctx2, s.svc2, s.ctrl2, envID)
		assert.NotNil(t, err)
	})
}

func newPipelineEnvironmentPayload(name string, envUUID uuid.UUID) *app.CreatePipelineEnvironmentsPayload {
	payload := &app.CreatePipelineEnvironmentsPayload{
		Data: &app.PipelineEnvironments{
			Name: name,
			Environments: []*app.EnvironmentAttributes{
				{EnvUUID: &envUUID},
			},
		},
	}
	return payload
}
