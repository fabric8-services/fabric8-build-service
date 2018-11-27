package controller_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/fabric8-services/fabric8-build/app"
	"github.com/fabric8-services/fabric8-build/app/test"
	"github.com/fabric8-services/fabric8-build/configuration"
	"github.com/fabric8-services/fabric8-build/controller"
	"github.com/fabric8-services/fabric8-build/gormapp"
	testauth "github.com/fabric8-services/fabric8-common/test/auth"
	testsuite "github.com/fabric8-services/fabric8-common/test/suite"
	"github.com/goadesign/goa"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
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
}

func TestPipelineEnvironmentController(t *testing.T) {
	config, err := configuration.New("")
	require.NoError(t, err)
	suite.Run(t, &PipelineEnvironmentControllerSuite{DBTestSuite: testsuite.NewDBTestSuite(config)})
}

func (s *PipelineEnvironmentControllerSuite) SetupSuite() {
	s.DBTestSuite.SetupSuite()

	s.db = gormapp.NewGormDB(s.DB)

	svc := testauth.UnsecuredService("ppl-test1")
	s.svc = svc

	svc2, err := testauth.ServiceAsUser("ppl-test2", testauth.NewIdentity())
	require.NoError(s.T(), err)
	s.svc2 = svc2

	s.ctx = s.svc.Context
	s.ctx2 = s.svc2.Context

	s.ctrl = controller.NewPipelineEnvironmentController(s.svc, s.db)
	s.ctrl2 = controller.NewPipelineEnvironmentController(s.svc2, s.db)

}

// createPipelineEnvironmentCtrlNoErroring we do this one manually cause the one from
// goatest one exit on errro without being able to catch
func (s *PipelineEnvironmentControllerSuite) createPipelineEnvironmentCtrlNoErroring() (*app.CreatePipelineEnvironmentsContext, *httptest.ResponseRecorder) {
	spaceID := uuid.NewV4()
	rw := httptest.NewRecorder()
	u := &url.URL{
		Path: fmt.Sprintf("/api/pipelines/environments/%v", spaceID),
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
	s.T().Run("ok", func(t *testing.T) {
		payload := newPipelineEnvironmentPayload("osio-stage-create", uuid.NewV4())
		_, newEnv := test.CreatePipelineEnvironmentsCreated(t, s.ctx2, s.svc2, s.ctrl2, uuid.NewV4(), payload)
		assert.NotNil(t, newEnv)
		assert.NotNil(t, newEnv.Data.ID)
		assert.NotNil(t, newEnv.Data.Environments[0].EnvUUID)

		// Same pipeline_name but different spaceID is OK
		payload = newPipelineEnvironmentPayload("osio-stage-create", uuid.NewV4())
		_, newEnv = test.CreatePipelineEnvironmentsCreated(t, s.ctx2, s.svc2, s.ctrl2, uuid.NewV4(), payload)
		assert.NotNil(t, newEnv)
		assert.NotNil(t, newEnv.Data.ID)
		assert.NotNil(t, newEnv.Data.Environments[0].EnvUUID)
	})

	s.T().Run("fail", func(t *testing.T) {
		conflict_space_id := uuid.NewV4()
		payload := newPipelineEnvironmentPayload("osio-stage-create-conflict", uuid.NewV4())

		_, newEnv := test.CreatePipelineEnvironmentsCreated(t, s.ctx2, s.svc2, s.ctrl2, conflict_space_id, payload)
		assert.NotNil(t, newEnv)

		response, err := test.CreatePipelineEnvironmentsConflict(t, s.ctx2, s.svc2, s.ctrl2, conflict_space_id, payload)
		require.NotNil(t, response.Header().Get("Location"))
		assert.Regexp(s.T(), ".*data_conflict_error.*", err.Errors)

		emptyPayload := &app.CreatePipelineEnvironmentsPayload{}
		createCtxerr, rw := s.createPipelineEnvironmentCtrlNoErroring()
		createCtxerr.Payload = emptyPayload
		jerr := s.ctrl2.Create(createCtxerr)
		require.Nil(t, jerr)
		require.Equal(t, 400, rw.Code)
	})

	s.T().Run("unauthorized", func(t *testing.T) {
		payload := newPipelineEnvironmentPayload("osio-stage", uuid.NewV4())
		_, err := test.CreatePipelineEnvironmentsUnauthorized(t, s.ctx, s.svc, s.ctrl, uuid.NewV4(), payload)
		assert.NotNil(t, err)
	})

}

func (s *PipelineEnvironmentControllerSuite) TestShow() {
	s.T().Run("ok", func(t *testing.T) {
		spaceID := uuid.NewV4()
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
