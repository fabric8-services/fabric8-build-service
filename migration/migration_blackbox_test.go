package migration_test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/fabric8-services/fabric8-build/configuration"
	"github.com/fabric8-services/fabric8-build/migration"
	"github.com/fabric8-services/fabric8-common/gormsupport"
	migrationsupport "github.com/fabric8-services/fabric8-common/migration"
	"github.com/fabric8-services/fabric8-common/resource"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	dbName          = "test"
	defaultHost     = "localhost"
	defaultPort     = "5433"
	defaultPassword = "mysecretpassword"
)

type MigrationTestSuite struct {
	suite.Suite
}

const (
	databaseName = "test"
)

var (
	sqlDB    *sql.DB
	host     string
	port     string
	password string
)

func TestMigration(t *testing.T) {
	suite.Run(t, new(MigrationTestSuite))
}

func (s *MigrationTestSuite) SetupTest() {
	resource.Require(s.T(), resource.Database)

	cfg, err := configuration.New("../config.yaml")
	if err != nil {
		panic(fmt.Errorf("Failed to setup the configuration: %s", err.Error()))
	}

	password = cfg.GetPostgresPassword()
	if password == "" {
		password = defaultPassword
	}

	host = os.Getenv("F8_POSTGRES_HOST")
	if host == "" {
		host = defaultHost
	}
	port = os.Getenv("F8_POSTGRES_PORT")
	if port == "" {
		port = defaultPort
	}

	dbConfig := fmt.Sprintf(
		"host=%s port=%s user=postgres password=%s sslmode=disable connect_timeout=5",
		host, port, password)

	db, err := sql.Open("postgres", dbConfig)
	require.NoError(s.T(), err, "cannot connect to database: %s", dbName)
	defer db.Close()

	_, err = db.Exec("DROP DATABASE IF EXISTS " + dbName)
	if err != nil && !gormsupport.IsInvalidCatalogName(err) {
		require.NoError(s.T(), err, "failed to drop database '%s'", dbName)
	}

	_, err = db.Exec("CREATE DATABASE " + dbName)
	require.NoError(s.T(), err, "failed to create database '%s'", dbName)
}

func (s *MigrationTestSuite) TestMigrate() {
	dbConfig := fmt.Sprintf("host=%s port=%s user=postgres password=%s dbname=%s sslmode=disable connect_timeout=5",
		host, port, password, dbName)
	var err error
	sqlDB, err = sql.Open("postgres", dbConfig)
	require.NoError(s.T(), err, "cannot connect to DB '%s'", dbName)
	defer sqlDB.Close()

	gormDB, err := gorm.Open("postgres", dbConfig)
	require.NoError(s.T(), err, "cannot connect to DB '%s'", dbName)
	defer gormDB.Close()
	s.T().Run("checkMigration001", checkMigration001)
}

func checkMigration001(t *testing.T) {
	err := migrationsupport.Migrate(sqlDB, databaseName, migration.Steps()[:2])
	require.NoError(t, err)

	pipeline_id := "80654c22-c378-40bc-a76e-33a4bcc45f79"
	t.Run("insert ok", func(t *testing.T) {
		_, err := sqlDB.Exec(`INSERT INTO pipelines (id, name, space_id) VALUES ('` +
			pipeline_id + `', 'pipeline1', uuid_generate_v4())`)
		require.NoError(t, err)
		_, err = sqlDB.Exec("INSERT INTO environments (environment_id, pipeline_id) VALUES (uuid_generate_v4(), '" +
			pipeline_id + "')")
		require.NoError(t, err)

	})
}
