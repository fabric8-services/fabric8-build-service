package application

import "github.com/fabric8-services/fabric8-build/build"

type Application interface {
	PipelineEnvMap() build.Repository
}

type Transaction interface {
	Application
	Commit() error
	Rollback() error
}

type DB interface {
	Application
	BeginTransaction() (Transaction, error)
}
