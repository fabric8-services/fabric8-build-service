package configuration_test

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/fabric8-services/fabric8-build-service/configuration"
	"github.com/goadesign/goa"
)

var reqLong *goa.RequestData
var reqShort *goa.RequestData
var config *configuration.Config

func TestMain(m *testing.M) {
	resetConfiguration()

	reqLong = &goa.RequestData{
		Request: &http.Request{Host: "api.service.domain.org"},
	}
	reqShort = &goa.RequestData{
		Request: &http.Request{Host: "api.domain.org"},
	}
	os.Exit(m.Run())
}

func resetConfiguration() {
	var err error

	// calling NewConfigurationData("") is same as GetConfigurationData()
	config, err = configuration.GetConfig()
	if err != nil {
		panic(fmt.Errorf("Failed to setup the configuration: %s", err.Error()))
	}
}
