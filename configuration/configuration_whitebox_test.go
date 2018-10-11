package configuration

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/fabric8-services/fabric8-auth/resource"
	"github.com/goadesign/goa"
	"github.com/stretchr/testify/assert"
)

var reqLong *goa.RequestData
var reqShort *goa.RequestData
var config *Config

func init() {

	// ensure that the content here is executed only once.
	reqLong = &goa.RequestData{
		Request: &http.Request{Host: "api.service.domain.org"},
	}
	reqShort = &goa.RequestData{
		Request: &http.Request{Host: "api.domain.org"},
	}
	resetConfiguration()
}

func resetConfiguration() {
	var err error
	config, err = GetConfig()
	if err != nil {
		panic(fmt.Errorf("Failed to setup the configuration: %s", err.Error()))
	}
}

func TestGetLogLevelOK(t *testing.T) {
	resource.Require(t, resource.UnitTest)

	key := "F8_LOG_LEVEL"
	realEnvValue := os.Getenv(key)

	err := os.Unsetenv(key)
	assert.Nil(t, err)
	defer func() {
		err := os.Setenv(key, realEnvValue)
		assert.Nil(t, err)
		resetConfiguration()
	}()

	assert.Equal(t, defaultLogLevel, config.GetLogLevel())

	err = os.Setenv(key, "warning")
	assert.Nil(t, err)
	resetConfiguration()

	assert.Equal(t, "warning", config.GetLogLevel())
}
