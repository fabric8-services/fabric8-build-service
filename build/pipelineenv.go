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

// Pipeline Env Map Structure
type PipelineEnvMap struct {
	gormsupport.Lifecycle
	ID           uuid.UUID  `sql:"type:uuid default uuid_generate_v4()" gorm:"primary_key"`
	Name         *string    `gorm:"not null;unique"` // Set field as not nullable and unique
	SpaceID      *uuid.UUID `sql:"type:uuid"`
	Environments []PipelineEnvironment
}

// Pipeline Environment contains entries of all PipelineEnvironmentMap-Environment associations
type PipelineEnvironment struct {
	gormsupport.Lifecycle
	EnvironmentID    *uuid.UUID `sql:"type:uuid"`
	PipelineEnvMapID uuid.UUID  `sql:"type:uuid" gorm:"column:pipelineenvmap_id"`
}

type Repository interface {
	Create(ctx context.Context, pipEnvMap *PipelineEnvMap) (*PipelineEnvMap, error)
	Load(ctx context.Context, ID uuid.UUID) (*PipelineEnvMap, error)
	List(ctx context.Context, spaceID uuid.UUID) ([]*PipelineEnvMap, error)
	Save(ctx context.Context, pipEnvMap *PipelineEnvMap) (*PipelineEnvMap, error)
}

type GormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{
		db: db,
	}
}

// Create a Pipeline Env Map
func (r *GormRepository) Create(ctx context.Context, pipEnvMap *PipelineEnvMap) (*PipelineEnvMap, error) {
	defer goa.MeasureSince([]string{"goa", "db", "pipeline_env_maps", "create"}, time.Now())

	err := r.db.Create(pipEnvMap).Error
	if err != nil {
		if gormsupport.IsUniqueViolation(err, "pipeline_env_maps_name_space_id_key") {
			return nil, errors.NewDataConflictError(fmt.Sprintf("pipeline_environment_map_name %s with spaceID %s already exists", *pipEnvMap.Name, *pipEnvMap.SpaceID))
		}

		log.Error(ctx, map[string]interface{}{"err": err},
			"unable to create pipeline-environment map")
		return nil, errs.WithStack(err)
	}

	return pipEnvMap, nil
}

// List all Pipeline Env Map in a space
func (r *GormRepository) List(ctx context.Context, spaceID uuid.UUID) ([]*PipelineEnvMap, error) {
	defer goa.MeasureSince([]string{"goa", "db", "pipeline_env_maps", "list"}, time.Now())
	var rows []*PipelineEnvMap
	tx := r.db.Model(&PipelineEnvMap{}).Where("space_id = ?", spaceID).Preload("Environments").Find(&rows)
	if tx.RecordNotFound() {
		log.Error(ctx, map[string]interface{}{"space_id": spaceID.String()},
			"state or known referer was empty")
		return nil, errors.NewNotFoundError("pipeline-environment", spaceID.String())
	}
	// This should not happen as I don't see what kind of other error (as long
	// schemas are created) than RecordNotFound can we have
	if tx.Error != nil {
		log.Error(ctx, map[string]interface{}{"err": tx.Error, "space_id": spaceID.String()},
			"unable to list the pipeline-environment by spaceID")
		return nil, errors.NewInternalError(ctx, tx.Error)
	}
	return rows, nil
}

// Load a Pipeline Env Map of given ID
func (r *GormRepository) Load(ctx context.Context, ID uuid.UUID) (*PipelineEnvMap, error) {
	defer goa.MeasureSince([]string{"goa", "db", "pipeline_env_maps", "load"}, time.Now())
	ppl := PipelineEnvMap{}
	tx := r.db.Model(&PipelineEnvMap{}).Where("id = ?", ID).Preload("Environments").First(&ppl)
	if tx.RecordNotFound() {
		log.Error(ctx, map[string]interface{}{"id": ID.String()},
			"state or known referer was empty")
		return nil, errors.NewNotFoundError("pipeline-environment", ID.String())
	}
	// This should not happen as I don't see what kind of other error (as long
	// schemas are created) than RecordNotFound can we have
	if tx.Error != nil {
		log.Error(ctx, map[string]interface{}{"err": tx.Error, "id": ID.String()},
			"unable to load the pipeline-environment by ID")
		return nil, errors.NewInternalError(ctx, tx.Error)
	}
	return &ppl, nil
}

// Save the given Pipeline Env Map
func (r *GormRepository) Save(ctx context.Context, p *PipelineEnvMap) (*PipelineEnvMap, error) {
	defer goa.MeasureSince([]string{"goa", "db", "pipeline_env_maps", "save"}, time.Now())
	ppl, err := r.Load(ctx, p.ID)
	if err != nil {
		log.Error(ctx, map[string]interface{}{
			"pipeline_environment_map_id": p.ID.String(),
			"err":                         err,
		}, "unable to load pipeline environment map")
		return nil, errors.NewInternalError(ctx, err)
	}

	tx := r.db.Model(ppl).Updates(p)
	if err := tx.Error; err != nil {
		if gormsupport.IsCheckViolation(tx.Error, "pipelineEnvMap_name_check") {
			return nil, errors.NewBadParameterError("Name", p.Name).Expected("not empty")
		}
		if gormsupport.IsUniqueViolation(tx.Error, "pipelineEnvMap_name_id") {
			return nil, errors.NewBadParameterError("Name", p.Name).Expected("unique")
		}
		log.Error(ctx, map[string]interface{}{
			"err":                         err,
			"pipeline_environment_map_id": p.ID,
		}, "unable to update pipeline environment map")
		return nil, errors.NewInternalError(ctx, err)
	}
	log.Info(ctx, map[string]interface{}{
		"pipelineEnvironment_id": p.ID,
	}, "pipelineEnvironment map updated successfully")
	return p, nil
}
