package env

import (
	"context"
	"github.com/fabric8-services/fabric8-build/application/env/envservice"
	"github.com/fabric8-services/fabric8-build/application/rest"
	"github.com/fabric8-services/fabric8-build/configuration"
	commonerr "github.com/fabric8-services/fabric8-common/errors"
	"github.com/fabric8-services/fabric8-common/goasupport"
	guuid "github.com/goadesign/goa/uuid"
	"github.com/pkg/errors"
	"github.com/prometheus/common/log"
	"net/http"
	"net/url"
)

type Environment struct {
	ID   guuid.UUID
	Name string
}

type ENVService interface {
	GetEnvList(ctx context.Context, spaceID string) (envs []Environment, e error)
}

type ENVServiceImpl struct {
	Config configuration.Config
	doer   rest.HttpDoer
}

// GetEnvList talks to the ENV service and return the list of env's in a given space
func (s *ENVServiceImpl) GetEnvList(ctx context.Context, spaceID string) (envs []Environment, e error) {
	remoteENVService, err := s.createClientWithContextSigner(ctx)
	if err != nil {
		return nil, err
	}

	spaceIDUUID, err := guuid.FromString(spaceID)
	if err != nil {
		return nil, err
	}

	res, err := remoteENVService.ListEnvironment(ctx, envservice.ListEnvironmentPath(spaceIDUUID))
	if err != nil {
		return nil, err
	}

	defer rest.CloseResponse(res)
	if res.StatusCode != http.StatusOK {
		bodyString := rest.ReadBody(res.Body)
		log.Error(ctx, map[string]interface{}{
			"spaceId":         spaceID,
			"response_status": res.Status,
			"response_body":   bodyString,
		}, "unable to get env list from ENV Service")
		if res.StatusCode == 401 {
			return nil, commonerr.NewUnauthorizedError("Not Authorized")
		} else {
			return nil, errors.Errorf("unable to get env list from ENV Service. Response status: %s. Response body: %s", res.Status, bodyString)
		}
	}

	envList, err := remoteENVService.DecodeEnvironmentsList(res)
	if err != nil {
		return nil, err
	}

	for _, env := range envList.Data {
		envs = append(envs, Environment{
			ID:   *env.ID,
			Name: *env.Attributes.Name,
		})
	}
	return envs, nil
}

// createClientWithContextSigner creates with a signer based on current context
func (s *ENVServiceImpl) createClientWithContextSigner(ctx context.Context) (*envservice.Client, error) {
	c, err := s.createClient()
	if err != nil {
		return nil, err
	}

	c.SetJWTSigner(goasupport.NewForwardSigner(ctx))
	return c, nil
}

func (s *ENVServiceImpl) createClient() (*envservice.Client, error) {
	envURL, e := s.Config.GetEnvServiceURL()
	if e != nil {
		return nil, e
	}
	u, err := url.Parse(envURL)
	if err != nil {
		return nil, err
	}

	c := envservice.New(s.doer)
	c.Host = u.Host
	c.Scheme = u.Scheme
	return c, nil
}
