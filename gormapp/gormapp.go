package gormapp

import (
	"fmt"
	"strconv"

	"github.com/fabric8-services/fabric8-build/application"
	"github.com/fabric8-services/fabric8-build/build"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

type TXIsoLevel int8

const (
	// See https://www.postgresql.org/docs/9.3/static/sql-set-transaction.html

	TXIsoLevelDefault TXIsoLevel = iota
	TXIsoLevelReadCommitted
	TXIsoLevelRepeatableRead
	TXIsoLevelSerializable
)

var _ application.DB = &GormDB{}

var _ application.Transaction = &GormTransaction{}

func NewGormDB(db *gorm.DB) *GormDB {
	g := new(GormDB)
	g.db = db.Set("gorm:association_autoupdate", true)
	g.txIsoLevel = ""
	return g
}

type GormBase struct {
	db *gorm.DB
}

type GormDB struct {
	GormBase
	txIsoLevel string
}

type GormTransaction struct {
	GormBase
}

func (g *GormBase) DB() *gorm.DB {
	return g.db
}

// See https://www.postgresql.org/docs/9.3/static/sql-set-transaction.html
func (g *GormDB) SetTransactionIsolationLevel(level TXIsoLevel) error {
	switch level {
	case TXIsoLevelReadCommitted:
		g.txIsoLevel = "READ COMMITTED"
	case TXIsoLevelRepeatableRead:
		g.txIsoLevel = "REPEATABLE READ"
	case TXIsoLevelSerializable:
		g.txIsoLevel = "SERIALIZABLE"
	case TXIsoLevelDefault:
		g.txIsoLevel = ""
	default:
		return fmt.Errorf("Unknown transaction isolation level: " + strconv.FormatInt(int64(level), 10))
	}
	return nil
}

func (g *GormDB) BeginTransaction() (application.Transaction, error) {
	tx := g.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	if len(g.txIsoLevel) != 0 {
		tx := tx.Exec(fmt.Sprintf("set transaction isolation level %s", g.txIsoLevel))
		if tx.Error != nil {
			return nil, tx.Error
		}
		return &GormTransaction{GormBase{tx}}, nil
	}
	return &GormTransaction{GormBase{tx}}, nil
}

func (g *GormTransaction) Commit() error {
	err := g.db.Commit().Error
	g.db = nil
	return errors.WithStack(err)
}

func (g *GormTransaction) Rollback() error {
	err := g.db.Rollback().Error
	g.db = nil
	return errors.WithStack(err)
}

func (g *GormBase) Pipeline() build.Repository {
	return build.NewRepository(g.db)
}
