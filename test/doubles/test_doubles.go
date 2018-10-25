package testdoubles

import (
	"log"
	"os"
	"testing"

	vcrrecorder "github.com/dnaeon/go-vcr/recorder"
	"github.com/fabric8-services/fabric8-build/auth"
	"github.com/fabric8-services/fabric8-build/configuration"
	"github.com/fabric8-services/fabric8-build/test/recorder"
	"github.com/stretchr/testify/require"
)

// LoadTestConfig load test config
func LoadTestConfig(t *testing.T) (*configuration.Config, func()) {
	reset := SetEnvironments(
		Env("F8_TEMPLATE_RECOMMENDER_EXTERNAL_NAME", "recommender.api.prod-preview.openshift.io"),
		Env("F8_TEMPLATE_RECOMMENDER_API_TOKEN", "xxxx"),
		Env("F8_TEMPLATE_DOMAIN", "d800.free-int.openshiftapps.com"))
	data, err := configuration.GetConfig()
	require.NoError(t, err)
	return data, reset
}

// Env return Environment env
func Env(key, value string) Environment {
	return Environment{key: key, value: value}
}

// Environment type
type Environment struct {
	key, value string
}

// SetEnvironments set all environements variable
func SetEnvironments(environments ...Environment) func() {
	originalValues := make([]Environment, 0, len(environments))

	for _, env := range environments {
		originalValues = append(originalValues, Env(env.key, os.Getenv(env.key)))
		err := os.Setenv(env.key, env.value)
		if err != nil {
			log.Fatal(err)
		}
	}
	return func() {
		for _, env := range originalValues {
			err := os.Setenv(env.key, env.value)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

// NewAuthService new auth service recorder
func NewAuthService(t *testing.T, cassetteFile, authURL string, options ...recorder.Option) (*auth.Service, func()) {
	authService, _, cleanup := NewAuthServiceWithRecorder(t, cassetteFile, authURL, options...)
	return authService, cleanup
}

// NewAuthServiceWithRecorder new auth service with recorder
func NewAuthServiceWithRecorder(t *testing.T, cassetteFile, authURL string, options ...recorder.Option) (*auth.Service, *vcrrecorder.Recorder, func()) {
	var clientOptions []configuration.HTTPClientOption
	var r *vcrrecorder.Recorder
	var err error
	if cassetteFile != "" {
		r, err = recorder.New(cassetteFile, options...)
		require.NoError(t, err)
		clientOptions = append(clientOptions, configuration.WithRoundTripper(r))
	}
	resetBack := SetEnvironments(Env("F8_AUTH_URL", authURL))
	config, err := configuration.GetConfig()
	require.NoError(t, err)

	authService := &auth.Service{
		Config:        config,
		ClientOptions: clientOptions,
	}
	return authService, r, func() {
		if r != nil {
			err := r.Stop()
			require.NoError(t, err)
		}
		resetBack()
	}
}
