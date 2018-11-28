package wit

import (
	"context"
	"net/http"
	"net/url"

	"github.com/fabric8-services/fabric8-build/application/rest"
	"github.com/fabric8-services/fabric8-build/application/wit/witservice"
	"github.com/fabric8-services/fabric8-build/configuration"
	commonerr "github.com/fabric8-services/fabric8-common/errors"
	"github.com/fabric8-services/fabric8-common/goasupport"
	"github.com/pkg/errors"

	guuid "github.com/goadesign/goa/uuid"
	"github.com/prometheus/common/log"
)

type Space struct {
	ID          guuid.UUID
	OwnerID     guuid.UUID
	Name        string
	Description string
}

type WITService interface {
	GetSpace(ctx context.Context, spaceID string) (space *Space, e error)
}

type WITServiceImpl struct {
	Config configuration.Config
	doer   rest.HttpDoer
}

// GetSpace talks to the WIT service to retrieve a space record for the specified spaceID, then returns space
func (s *WITServiceImpl) GetSpace(ctx context.Context, spaceID string) (space *Space, e error) {
	remoteWITService, err := s.createClientWithContextSigner(ctx)
	if err != nil {
		return nil, err
	}

	spaceIDUUID, err := guuid.FromString(spaceID)
	if err != nil {
		return nil, err
	}

	res, err := remoteWITService.ShowSpace(ctx, witservice.ShowSpacePath(spaceIDUUID), nil, nil)
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
		}, "unable to get space from WIT")
		if res.StatusCode == 404 {
			return nil, commonerr.NewNotFoundErrorFromString("Cannot find space: " + spaceID)
		} else {
			return nil, errors.Errorf("unable to get space from WIT. Response status: %s. Response body: %s", res.Status, bodyString)
		}
	}

	spaceSingle, err := remoteWITService.DecodeSpaceSingle(res)
	if err != nil {
		return nil, err
	}

	return &Space{
		ID:          *spaceSingle.Data.ID,
		Name:        *spaceSingle.Data.Attributes.Name,
		Description: *spaceSingle.Data.Attributes.Description,
		OwnerID:     *spaceSingle.Data.Relationships.OwnedBy.Data.ID}, nil
}

// createClientWithContextSigner creates with a signer based on current context
func (s *WITServiceImpl) createClientWithContextSigner(ctx context.Context) (*witservice.Client, error) {
	c, err := s.createClient()
	if err != nil {
		return nil, err
	}

	c.SetJWTSigner(goasupport.NewForwardSigner(ctx))
	return c, nil
}

func (s *WITServiceImpl) createClient() (*witservice.Client, error) {
	witURL, e := s.Config.GetWITURL()
	if e != nil {
		return nil, e
	}
	u, err := url.Parse(witURL)
	if err != nil {
		return nil, err
	}

	c := witservice.New(s.doer)
	c.Host = u.Host
	c.Scheme = u.Scheme
	return c, nil
}
