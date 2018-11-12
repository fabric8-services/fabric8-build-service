package controller_test

import (
	"context"
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

	// TODO(chmouel): change this when we have jwt support,
	svc := testauth.UnsecuredService("ppl-test1")
	s.svc = svc
	s.ctx = s.svc.Context
	s.ctrl = controller.NewPipelineEnvironmentController(s.svc, s.db)
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

		// TODO(chmouel): add this when we have show controller
		// _, env := test.ShowPipelineEnvironmentsOK(t, s.ctx, s.svc, s.ctrl, *newEnv.Data.ID)
		// require.NotNil(t, env)
		// assert.Equal(t, env.Data.ID, newEnv.Data.ID)
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
