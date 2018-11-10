package migration

import (
	"database/sql"

	"github.com/fabric8-services/fabric8-common/migration"
)

func Migrate(db *sql.DB, catalog string) error {
	return migration.Migrate(db, catalog, Steps())
}

type Scripts [][]string

func Steps() Scripts {
	return [][]string{
		{"000-bootstrap.sql"},
		{"001-pipelineenv.sql"},
	}
}

func (s Scripts) Asset(name string) ([]byte, error) {
	return Asset(name)
}

func (s Scripts) AssetNameWithArgs() [][]string {
	return s
}
