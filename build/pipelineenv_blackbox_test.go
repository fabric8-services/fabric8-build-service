package build_test

import (
	"context"
	"testing"

	"github.com/fabric8-services/fabric8-build/configuration"
	testsuite "github.com/fabric8-services/fabric8-common/test/suite"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/fabric8-services/fabric8-build/build"
)

type BuildRepositorySuite struct {
	testsuite.DBTestSuite
	buildRepo *build.GormRepository
}

func TestBuildRepository(t *testing.T) {
	config, err := configuration.New("")
	require.NoError(t, err)
	suite.Run(t, &BuildRepositorySuite{DBTestSuite: testsuite.NewDBTestSuite(config)})
}

func (s *BuildRepositorySuite) SetupSuite() {
	s.DBTestSuite.SetupSuite()
	s.buildRepo = build.NewRepository(s.DB)
}

func (s *BuildRepositorySuite) TestCreate() {
	spaceID, envUUID := uuid.NewV4(), uuid.NewV4()
	np := newPipelineEnvMap("pipeline1", spaceID, envUUID)
	ppl, err := s.buildRepo.Create(context.Background(), np)

	require.NoError(s.T(), err)
	require.NotNil(s.T(), ppl)

	// Test that auto associations is done
	assert.Equal(s.T(), 1, len(ppl.Environments))
	assert.Equal(s.T(), ppl.ID, ppl.Environments[0].PipelineEnvMapID)

	// Test unique constraint violation
	ppl, err = s.buildRepo.Create(context.Background(), np)
	require.Error(s.T(), err)
	require.Nil(s.T(), ppl)
	assert.Regexp(s.T(), ".*duplicate key value violates unique constraint.*", err.Error())

	// Test empty name is a failure
	npe := &build.PipelineEnvMap{}
	ppl, err = s.buildRepo.Create(context.Background(), npe)
	require.Error(s.T(), err)
	require.EqualError(s.T(), err, "pq: null value in column \"name\" violates not-null constraint")
	require.Nil(s.T(), ppl)
}

func (s *BuildRepositorySuite) TestShow() {
	spaceID, envUUID := uuid.NewV4(), uuid.NewV4()
	newPipEnvMap, err := s.buildRepo.Create(context.Background(), newPipelineEnvMap("pipelineShow", spaceID, envUUID))
	require.NoError(s.T(), err)
	require.NotNil(s.T(), newPipEnvMap)

	pipEnvMap, err := s.buildRepo.Load(context.Background(), newPipEnvMap.ID)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), pipEnvMap)
	assert.Equal(s.T(), newPipEnvMap.ID, pipEnvMap.ID)
}

func (s *BuildRepositorySuite) TestList() {
	spaceID, envUUID, envUUID2 := uuid.NewV4(), uuid.NewV4(), uuid.NewV4()
	newEnv, err := s.buildRepo.Create(context.Background(), newPipelineEnvMap("pipelineShow", spaceID, envUUID))
	newEnv2, err2 := s.buildRepo.Create(context.Background(), newPipelineEnvMap("pipelineShow2", spaceID, envUUID2))
	require.NoError(s.T(), err)
	require.NoError(s.T(), err2)
	require.NotNil(s.T(), newEnv)
	require.NotNil(s.T(), newEnv2)

	env, err := s.buildRepo.List(context.Background(), spaceID)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), env)
	assert.Equal(s.T(), 2, len(env))

	env2, err2 := s.buildRepo.List(context.Background(), uuid.NewV4())
	require.NoError(s.T(), err2)
	assert.NotNil(s.T(), env2)
	assert.Equal(s.T(), 0, len(env2))
}

func (s *BuildRepositorySuite) TestSave() {
	spaceID, envUUID, envUUID2, envUUID3 := uuid.NewV4(), uuid.NewV4(), uuid.NewV4(), uuid.NewV4()
	pipeline := newPipelineEnvMap("pipelineShow", spaceID, envUUID)
	newEnv, err := s.buildRepo.Create(context.Background(), pipeline)
	newEnv2, err2 := s.buildRepo.Create(context.Background(), newPipelineEnvMap("pipelineShow2", spaceID, envUUID2))
	require.NoError(s.T(), err)
	require.NoError(s.T(), err2)
	require.NotNil(s.T(), newEnv)
	require.NotNil(s.T(), newEnv2)

	pipelineUpdate := updatePipelineEnvMap(pipeline, envUUID3)
	env, err := s.buildRepo.Save(context.Background(), pipelineUpdate)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), env)
	require.NotNil(s.T(), env)
	require.NotNil(s.T(), env)
	assert.Equal(s.T(), envUUID3, *(env.Environments[0].EnvironmentID))
}

func newPipelineEnvMap(name string, spaceID, envUUID uuid.UUID) *build.PipelineEnvMap {
	ppl := &build.PipelineEnvMap{
		Name:    &name,
		SpaceID: &spaceID,
		Environments: []build.PipelineEnvironment{
			{EnvironmentID: &envUUID},
		},
	}
	return ppl
}

func updatePipelineEnvMap(pipEnvMap *build.PipelineEnvMap, envUUID uuid.UUID) *build.PipelineEnvMap {
	ppl := &build.PipelineEnvMap{
		Name:    pipEnvMap.Name,
		SpaceID: pipEnvMap.SpaceID,
		ID:      pipEnvMap.ID,
		Environments: []build.PipelineEnvironment{
			{EnvironmentID: &envUUID},
		},
	}
	return ppl
}
