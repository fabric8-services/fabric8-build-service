package build_test

import (
	"context"
	"log"
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

	//TODO(chmouel): really need something better
	for _, table := range []string{"environments", "pipelines"} {
		_, err := s.DB.DB().Exec("DELETE FROM " + table + " CASCADE")
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (s *BuildRepositorySuite) TestCreate() {
	spaceID := uuid.NewV4()
	envUUID := uuid.NewV4()
	np := newPipeline("pipeline1", spaceID, envUUID)
	ppl, err := s.buildRepo.Create(context.Background(), np)

	require.NoError(s.T(), err)
	require.NotNil(s.T(), ppl)

	// Test that auto associations is done
	assert.Equal(s.T(), 1, len(ppl.Environment))
	assert.Equal(s.T(), ppl.ID, ppl.Environment[0].PipelineID)

	// Test unique constraint violation
	ppl, err = s.buildRepo.Create(context.Background(), np)
	require.Error(s.T(), err)
	require.Nil(s.T(), ppl)
	assert.Regexp(s.T(), ".*duplicate key value violates unique constraint.*", err.Error())

	// Test empty name is a failure
	npe := &build.Pipeline{}
	ppl, err = s.buildRepo.Create(context.Background(), npe)
	require.Error(s.T(), err)
	require.EqualError(s.T(), err, "pq: null value in column \"name\" violates not-null constraint")
	require.Nil(s.T(), ppl)

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
