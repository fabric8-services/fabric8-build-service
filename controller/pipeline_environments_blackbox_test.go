package controller_test

import (
	"context"
	"fmt"
	"log"
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

	ctrl     *controller.PipelineEnvironmentController
	ctrl2    *controller.PipelineEnvironmentController
	prodCtrl *controller.PipelineEnvironmentController
}

func TestEnvironmentController(t *testing.T) {
	config, err := configuration.New("")
	require.NoError(t, err)
	suite.Run(t, &PipelineEnvironmentControllerSuite{DBTestSuite: testsuite.NewDBTestSuite(config)})
}

func (s *PipelineEnvironmentControllerSuite) SetupSuite() {
	s.DBTestSuite.SetupSuite()

	s.db = gormapp.NewGormDB(s.DB)

	//TODO(chmouel): really need something better
	for _, table := range []string{"environments", "pipelines"} {
		_, err := s.DB.DB().Exec("DELETE FROM " + table + " CASCADE")
		if err != nil {
			log.Fatal(err)
		}
	}

	// TODO(chmouel): change this when we have jwt support,
	svc := testauth.UnsecuredService("ppl-test1")
	s.svc = svc
	s.ctx = s.svc.Context
	s.ctrl = controller.NewPipelineEnvironmentController(s.svc, s.db)
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
	goaCtx := goa.NewContext(goa.WithAction(s.ctx, "PipelineEnvironmentsTest"), rw, req, prms)
	createCtx, __err := app.NewCreatePipelineEnvironmentsContext(goaCtx, req, s.svc)
	if __err != nil {
		panic("invalid test data " + __err.Error()) // bug
	}
	return createCtx, rw
}

func (s *PipelineEnvironmentControllerSuite) TestCreate() {
	s.T().Run("ok", func(t *testing.T) {
		spaceID := uuid.NewV4()
		envID := uuid.NewV4()

		payload := newPipelineEnvironmentPayload("osio-stage", envID)
		_, newEnv := test.CreatePipelineEnvironmentsCreated(t, s.ctx, s.svc, s.ctrl, spaceID, payload)
		assert.NotNil(t, newEnv)
		assert.NotNil(t, newEnv.Data.ID)
		assert.NotNil(t, newEnv.Data.Environments[0].EnvUUID)

		createCtxerr, rw := s.createPipelineEnvironmentCtrlNoErroring()
		createCtxerr.Payload = payload
		s.ctrl.Create(createCtxerr)
		require.Equal(s.T(), 500, rw.Code)

		createCtxerr, rw = s.createPipelineEnvironmentCtrlNoErroring()
		emptyPayload := &app.CreatePipelineEnvironmentsPayload{}
		createCtxerr.Payload = emptyPayload
		s.ctrl.Create(createCtxerr)
		require.Equal(s.T(), 400, rw.Code)
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
