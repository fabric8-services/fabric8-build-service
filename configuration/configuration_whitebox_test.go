package configuration

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/fabric8-services/fabric8-auth/resource"
	"github.com/goadesign/goa"
	"github.com/stretchr/testify/assert"
)

var reqLong *goa.RequestData
var reqShort *goa.RequestData
var config *Config

const (
	envF8DevMode                    = "F8_DEVELOPER_MODE_ENABLED"
	envF8LogJSON                    = "F8_LOG_JSON"
	envF8DiagnoseHTTPAddresse       = "F8_DIAGNOSE_HTTP_ADDRESS"
	envF8Environment                = "F8_ENVIRONMENT"
	envF8AuthURL                    = "F8_AUTH_URL"
	envF8PostgresTransactionTimeout = "F8_POSTGRES_TRANSACTION_TIMEOUT"
)

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

func TestConfigFileNotFound(t *testing.T) {
	resource.Require(t, resource.UnitTest)
	_, err := New("/unkown/file")
	assert.Error(t, err)
}

func TestDeveloperEnabled(t *testing.T) {
	resource.Require(t, resource.UnitTest)

	realEnvValue := os.Getenv(envF8DevMode)

	os.Setenv(envF8DevMode, "1")
	cfg, _ := New("")
	assert.True(t, cfg.DeveloperModeEnabled())

	os.Unsetenv(envF8DevMode)
	cfg, _ = New("")
	assert.False(t, cfg.DeveloperModeEnabled())

	os.Setenv(envF8DevMode, realEnvValue)
}

func TestGetAuthServiceURL(t *testing.T) {
	resource.Require(t, resource.UnitTest)

	realEnvValue := os.Getenv(envF8AuthURL)
	realDevValue := os.Getenv(envF8DevMode)

	os.Setenv(envF8AuthURL, "https://test.openshift.io")
	cfg, _ := New("")
	assert.Equal(t, "https://test.openshift.io", cfg.GetAuthServiceURL())

	os.Unsetenv(envF8AuthURL)
	os.Unsetenv(envF8DevMode)
	cfg, _ = New("")
	assert.Equal(t, "http://localhost:8089", cfg.GetAuthServiceURL())

	os.Unsetenv(envF8AuthURL)
	os.Setenv(envF8DevMode, "true")
	cfg, _ = New("")
	assert.Equal(t, "https://auth.prod-preview.openshift.io", cfg.GetAuthServiceURL())

	os.Setenv(envF8DevMode, realDevValue)
	os.Setenv(envF8AuthURL, realEnvValue)
}

func TestGetEnvironment(t *testing.T) {
	resource.Require(t, resource.UnitTest)

	realEnvValue := os.Getenv(envF8Environment)

	os.Setenv(envF8Environment, "ENVIRON")
	cfg, _ := New("")
	assert.Equal(t, "ENVIRON", cfg.GetEnvironment())

	os.Unsetenv(envF8Environment)
	cfg, _ = New("")
	assert.Equal(t, "local", cfg.GetEnvironment())

	os.Setenv(envF8Environment, realEnvValue)
}

func TestLogJSONConfig(t *testing.T) {
	resource.Require(t, resource.UnitTest)

	realDevEnvValue := os.Getenv(envF8DevMode)

	realLogJSONEnvValue := os.Getenv(envF8LogJSON)

	os.Unsetenv(envF8DevMode)
	os.Unsetenv(envF8LogJSON)
	cfg, _ := New("")
	assert.True(t, cfg.IsLogJSON())

	os.Setenv(envF8DevMode, "1")
	os.Unsetenv(envF8LogJSON)
	cfg, _ = New("")
	assert.False(t, cfg.IsLogJSON())

	os.Setenv(envF8DevMode, "1")
	os.Setenv(envF8LogJSON, "1")
	cfg, _ = New("")
	assert.True(t, cfg.IsLogJSON())

	os.Setenv(envF8DevMode, realDevEnvValue)
	os.Setenv(envF8LogJSON, realLogJSONEnvValue)
}

func TestIsDiagnosisAddress(t *testing.T) {
	resource.Require(t, resource.UnitTest)

	realDevEnvValue := os.Getenv(envF8DevMode)

	realDiagEnvValue := os.Getenv(envF8DiagnoseHTTPAddresse)

	os.Unsetenv(envF8DevMode)
	os.Setenv(envF8DiagnoseHTTPAddresse, "FOO")
	cfg, _ := New("")
	assert.Equal(t, "FOO", cfg.GetDiagnoseHTTPAddress())

	os.Setenv(envF8DevMode, "1")
	os.Unsetenv(envF8DiagnoseHTTPAddresse)
	cfg, _ = New("")
	assert.Equal(t, "127.0.0.1:0", cfg.GetDiagnoseHTTPAddress())

	os.Unsetenv(envF8DevMode)
	os.Unsetenv(envF8DiagnoseHTTPAddresse)
	cfg, _ = New("")
	assert.Equal(t, "", cfg.GetDiagnoseHTTPAddress())

	os.Setenv(envF8DevMode, realDevEnvValue)
	os.Setenv(envF8DiagnoseHTTPAddresse, realDiagEnvValue)
}

func TestGetTransactionTimeoutOK(t *testing.T) {
	resource.Require(t, resource.UnitTest)

	realEnvValue := os.Getenv(envF8PostgresTransactionTimeout)

	os.Unsetenv(envF8PostgresTransactionTimeout)
	defer func() {
		os.Setenv(envF8PostgresTransactionTimeout, realEnvValue)
		resetConfiguration()
	}()

	assert.Equal(t, time.Duration(5*time.Minute), config.GetPostgresTransactionTimeout())

	os.Setenv(envF8PostgresTransactionTimeout, "6m")
	resetConfiguration()

	assert.Equal(t, time.Duration(6*time.Minute), config.GetPostgresTransactionTimeout())
}
