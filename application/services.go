package application

import (
	"github.com/fabric8-services/fabric8-build/application/env"
	"github.com/fabric8-services/fabric8-build/application/wit"
	"github.com/fabric8-services/fabric8-build/configuration"
)

type ServiceFactory interface {
	WITService() wit.WITService
	ENVService() env.ENVService
}

type serviceFactoryImpl struct {
	Config *configuration.Config
}

func (s serviceFactoryImpl) WITService() wit.WITService {
	return &wit.WITServiceImpl{
		Config: *s.Config,
	}
}

func (s serviceFactoryImpl) ENVService() env.ENVService {
	return &env.ENVServiceImpl{
		Config: *s.Config,
	}
}

func NewServiceFactory(config *configuration.Config) ServiceFactory {
	return serviceFactoryImpl{
		Config: config,
	}
}
