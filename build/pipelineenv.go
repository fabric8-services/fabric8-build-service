package build

import (
	"context"
	"fmt"
	"time"

	"github.com/fabric8-services/fabric8-common/errors"
	"github.com/fabric8-services/fabric8-common/gormsupport"
	"github.com/goadesign/goa"
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
	Load(ctx context.Context, spaceID uuid.UUID) (*Pipeline, error)
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
	defer goa.MeasureSince([]string{"goa", "db", "pipeline", "create"}, time.Now())

	err := r.db.Create(pipl).Error
	if err != nil {
		if gormsupport.IsUniqueViolation(err, "pipelines_name_space_id_key") {
			return nil, errors.NewDataConflictError(fmt.Sprintf("pipeline_name %s with spaceID %s already exists", *pipl.Name, pipl.SpaceID))
		}

		log.Error(ctx, map[string]interface{}{"err": err},
			"unable to create pipeline")
		return nil, errs.WithStack(err)
	}

	return pipl, nil
}

func (r *GormRepository) Load(ctx context.Context, spaceID uuid.UUID) (*Pipeline, error) {
	defer goa.MeasureSince([]string{"goa", "db", "pipeline", "load"}, time.Now())
	ppl := Pipeline{}
	tx := r.db.Model(&Pipeline{}).Where("space_id = ?", spaceID).Preload("Environment").First(&ppl)
	if tx.RecordNotFound() {
		log.Error(ctx, map[string]interface{}{"space_id": spaceID.String()},
			"state or known referer was empty")
		return nil, errors.NewNotFoundError("pipeline", spaceID.String())
	}
	// This should not happen as I don't see what kind of other error (as long
	// schemas are created) than RecordNotFound can we have
	if tx.Error != nil {
		log.Error(ctx, map[string]interface{}{"err": tx.Error, "space_id": spaceID.String()},
			"unable to load the pipeline by spaceID")
		return nil, errors.NewInternalError(ctx, tx.Error)
	}
	return &ppl, nil
}
