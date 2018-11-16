package build

import (
	"context"

	"github.com/fabric8-services/fabric8-common/gormsupport"
	"github.com/jinzhu/gorm"
	errs "github.com/pkg/errors"
	"github.com/prometheus/common/log"
	uuid "github.com/satori/go.uuid"
)

type Pipeline struct {
	gormsupport.Lifecycle
	ID          uuid.UUID  `sql:"type:uuid default uuid_generate_v4()" gorm:"primary_key"`
	Name        *string    `gorm:"not null;unique"` // Set field as not nullable and unique
	SpaceID     *uuid.UUID `sql:"type:uuid"`
	Environment []Environment
}

type Environment struct {
	gormsupport.Lifecycle
	EnvironmentID *uuid.UUID `sql:"type:uuid"`
	PipelineID    uuid.UUID  `sql:"type:uuid"`
	ID            uuid.UUID  `sql:"type:uuid default uuid_generate_v4()" gorm:"primary_key"`
}

type Repository interface {
	Create(ctx context.Context, pipl *Pipeline) (*Pipeline, error)
}

type GormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{
		db: db,
	}
}

func (r *GormRepository) Create(ctx context.Context, pipl *Pipeline) (*Pipeline, error) {
	err := r.db.Create(pipl).Error
	if err != nil {
		log.Error(ctx, map[string]interface{}{"err": err},
			"unable to create pipeline")
		return nil, errs.WithStack(err)
	}

	return pipl, nil
}
