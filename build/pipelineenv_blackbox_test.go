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
	spaceID := uuid.NewV4()
	envUUID := uuid.NewV4()
	newPipeline := newPipeline("pipeline1", spaceID, envUUID)
	ppl, err := s.buildRepo.Create(context.Background(), newPipeline)

	require.NoError(s.T(), err)
	require.NotNil(s.T(), ppl)

	// Test that auto associations is done
	assert.Equal(s.T(), 1, len(ppl.Environment))
	assert.Equal(s.T(), ppl.ID, ppl.Environment[0].PipelineID)
}

func newPipeline(name string, spaceID, envUUID uuid.UUID) *build.Pipeline {
	ppl := &build.Pipeline{
		Name:    &name,
		SpaceID: &spaceID,
		Environment: []build.Environment{
			{EnvironmentID: &envUUID},
		},
	}
	return ppl
}
