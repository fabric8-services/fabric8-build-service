package testdoubles

import (
	"os"
	"testing"

	vcrrecorder "github.com/dnaeon/go-vcr/recorder"
	"github.com/fabric8-services/fabric8-build-service/auth"
	"github.com/fabric8-services/fabric8-build-service/configuration"
	"github.com/fabric8-services/fabric8-build-service/test/recorder"
	"github.com/stretchr/testify/require"
)

func LoadTestConfig(t *testing.T) (*configuration.Config, func()) {
	reset := SetEnvironments(
		Env("F8_TEMPLATE_RECOMMENDER_EXTERNAL_NAME", "recommender.api.prod-preview.openshift.io"),
		Env("F8_TEMPLATE_RECOMMENDER_API_TOKEN", "xxxx"),
		Env("F8_TEMPLATE_DOMAIN", "d800.free-int.openshiftapps.com"))
	data, err := configuration.GetConfig()
	require.NoError(t, err)
	return data, reset
}

func Env(key, value string) Environment {
	return Environment{key: key, value: value}
}

type Environment struct {
	key, value string
}

func SetEnvironments(environments ...Environment) func() {
	originalValues := make([]Environment, 0, len(environments))

	for _, env := range environments {
		originalValues = append(originalValues, Env(env.key, os.Getenv(env.key)))
		os.Setenv(env.key, env.value)
	}
	return func() {
		for _, env := range originalValues {
			os.Setenv(env.key, env.value)
		}
	}
}

func NewAuthService(t *testing.T, cassetteFile, authURL string, options ...recorder.Option) (*auth.Service, func()) {
	authService, _, cleanup := NewAuthServiceWithRecorder(t, cassetteFile, authURL, options...)
	return authService, cleanup
}

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
