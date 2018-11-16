package configuration

import (
	"fmt"
	"os"
	"strings"
	"time"

	errs "github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	defaultLogLevel         = "info"
	varCleanTestDataEnabled = "clean.test.data"
	varDBLogsEnabled        = "enable.db.logs"
	varDeveloperModeEnabled = "developer.mode.enabled"
	varDiagnoseHTTPAddress  = "diagnose.http.address"
	varEnvironment          = "environment"
	varHTTPAddress          = "http.address"
	varLogJSON              = "log.json"
	varLogLevel             = "log.level"
	varMetricsHTTPAddress   = "metrics.http.address"
	varSentryDSN            = "sentry.dsn"

	// Postgres
	varPostgresHost                 = "postgres.host"
	varPostgresPort                 = "postgres.port"
	varPostgresUser                 = "postgres.user"
	varPostgresDatabase             = "postgres.database"
	varPostgresPassword             = "postgres.password"
	varPostgresSSLMode              = "postgres.sslmode"
	varPostgresConnectionTimeout    = "postgres.connection.timeout"
	varPostgresTransactionTimeout   = "postgres.transaction.timeout"
	varPostgresConnectionRetrySleep = "postgres.connection.retrysleep"
	varPostgresConnectionMaxIdle    = "postgres.connection.maxidle"
	varPostgresConnectionMaxOpen    = "postgres.connection.maxopen"

	varAuthURL      = "auth.url"
	varAuthClientID = "service.account.id"
	varClientSecret = "service.account.secret"
	varAuthTokenKey = "auth.token.key"
)

// New creates a configuration reader object using a configurable configuration
// file path.
func New(configFilePath string) (*Config, error) {
	c := Config{
		v: viper.New(),
	}
	c.v.SetEnvPrefix("F8")
	c.v.AutomaticEnv()
	c.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	c.v.SetTypeByDefaultValue(true)
	c.setConfigDefaults()

	if configFilePath != "" {
		c.v.SetConfigType("yaml")
		c.v.SetConfigFile(configFilePath)
		err := c.v.ReadInConfig() // Find and read the config file
		if err != nil {           // Handle errors reading the config file
			return nil, errs.Errorf("Fatal error config file: %s \n", err)
		}
	}
	return &c, nil
}

// Config encapsulates the Viper configuration registry which stores the
// configuration data in-memory.
type Config struct {
	v *viper.Viper
}

// GetConfig is a wrapper over NewConfigurationData which reads configuration file path
// from the environment variable.
func GetConfig() (*Config, error) {
	return New(getMainConfigFile())
}

func getMainConfigFile() string {
	// This was either passed as a env var or set inside main.go from --config
	envConfigPath, _ := os.LookupEnv("BUILD_CONFIG_FILE_PATH")
	return envConfigPath
}

func (c *Config) setConfigDefaults() {
	c.v.SetTypeByDefaultValue(true)

	c.v.SetDefault(varLogLevel, defaultLogLevel)
	c.v.SetDefault(varHTTPAddress, "0.0.0.0:8080")
	c.v.SetDefault(varMetricsHTTPAddress, "0.0.0.0:8080")
	c.v.SetDefault(varDeveloperModeEnabled, false)
	c.v.SetDefault(varCleanTestDataEnabled, true)
	c.v.SetDefault(varDBLogsEnabled, false)

	//---------
	// Postgres
	//---------
	c.v.SetDefault(varPostgresHost, "localhost")
	c.v.SetDefault(varPostgresPort, 5432)
	c.v.SetDefault(varPostgresUser, "postgres")
	c.v.SetDefault(varPostgresDatabase, "postgres")
	c.v.SetDefault(varPostgresPassword, "mysecretpassword")
	c.v.SetDefault(varPostgresSSLMode, "disable")
	c.v.SetDefault(varPostgresConnectionTimeout, 5)
	c.v.SetDefault(varPostgresConnectionMaxIdle, -1)
	c.v.SetDefault(varPostgresConnectionMaxOpen, -1)
	// Number of seconds to wait before trying to connect again
	c.v.SetDefault(varPostgresConnectionRetrySleep, time.Second)

	// Timeout of a transaction in minutes
	c.v.SetDefault(varPostgresTransactionTimeout, 5*time.Minute)

	// Auth
	c.v.SetDefault(varAuthURL, "http://localhost:8089")
	c.v.SetDefault(varAuthClientID, "341c283f-0cd7-48a8-9281-4583aceb3617") // TODO: This is using idler creds, add this service example to auth
	c.v.SetDefault(varClientSecret, "secret")
}

// DeveloperModeEnabled returns `true` if development related features (as set via default, config file, or environment variable),
// e.g. token generation endpoint are enabled
func (c *Config) DeveloperModeEnabled() bool {
	return c.v.GetBool(varDeveloperModeEnabled)
}

// GetEnvironment returns the current environment application is deployed in
// like 'production', 'prod-preview', 'local', etc as the value of environment variable
// `F8_ENVIRONMENT` is set.
func (c *Config) GetEnvironment() string {
	if c.v.IsSet(varEnvironment) {
		return c.v.GetString(varEnvironment)
	}
	return "local"
}

// IsLogJSON returns if we should log json format (as set via config file or environment variable)
func (c *Config) IsLogJSON() bool {
	if c.v.IsSet(varLogJSON) {
		return c.v.GetBool(varLogJSON)
	}
	if c.DeveloperModeEnabled() {
		return false
	}
	return true
}

// GetHTTPAddress returns the HTTP address (as set via default, config file, or environment variable)
// that the wit server binds to (e.g. "0.0.0.0:8080")
func (c *Config) GetHTTPAddress() string {
	return c.v.GetString(varHTTPAddress)
}

// GetMetricsHTTPAddress returns the address the /metrics endpoing will be mounted.
// By default GetMetricsHTTPAddress is the same as GetHTTPAddress
func (c *Config) GetMetricsHTTPAddress() string {
	return c.v.GetString(varMetricsHTTPAddress)
}

// GetDiagnoseHTTPAddress returns the address of where to start the gops handler.
// By default GetDiagnoseHTTPAddress is 127.0.0.1:0 in devMode, but turned off in prod mode
// unless explicitly configured
func (c *Config) GetDiagnoseHTTPAddress() string {
	if c.v.IsSet(varDiagnoseHTTPAddress) {
		return c.v.GetString(varDiagnoseHTTPAddress)
	} else if c.DeveloperModeEnabled() {
		return "127.0.0.1:0"
	}
	return ""
}

// GetLogLevel returns the loggging level (as set via config file or environment variable)
func (c *Config) GetLogLevel() string {
	return c.v.GetString(varLogLevel)
}

// GetPostgresHost returns the postgres host as set via default, config file, or environment variable
func (c *Config) GetPostgresHost() string {
	return c.v.GetString(varPostgresHost)
}

// GetPostgresPort returns the postgres port as set via default, config file, or environment variable
func (c *Config) GetPostgresPort() int64 {
	return c.v.GetInt64(varPostgresPort)
}

// GetPostgresUser returns the postgres user as set via default, config file, or environment variable
func (c *Config) GetPostgresUser() string {
	return c.v.GetString(varPostgresUser)
}

// GetPostgresDatabase returns the postgres database as set via default, config file, or environment variable
func (c *Config) GetPostgresDatabase() string {
	return c.v.GetString(varPostgresDatabase)
}

// GetPostgresPassword returns the postgres password as set via default, config file, or environment variable
func (c *Config) GetPostgresPassword() string {
	return c.v.GetString(varPostgresPassword)
}

// GetPostgresSSLMode returns the postgres sslmode as set via default, config file, or environment variable
func (c *Config) GetPostgresSSLMode() string {
	return c.v.GetString(varPostgresSSLMode)
}

// GetPostgresConnectionTimeout returns the postgres connection timeout as set via default, config file, or environment variable
func (c *Config) GetPostgresConnectionTimeout() int64 {
	return c.v.GetInt64(varPostgresConnectionTimeout)
}

// GetPostgresConnectionRetrySleep returns the number of seconds (as set via default, config file, or environment variable)
// to wait before trying to connect again
func (c *Config) GetPostgresConnectionRetrySleep() time.Duration {
	return c.v.GetDuration(varPostgresConnectionRetrySleep)
}

// GetPostgresTransactionTimeout returns the number of minutes to timeout a transaction
func (c *Config) GetPostgresTransactionTimeout() time.Duration {
	return c.v.GetDuration(varPostgresTransactionTimeout)
}

// GetPostgresConnectionMaxIdle returns the number of connections that should be keept alive in the database connection pool at
// any given time. -1 represents no restrictions/default behavior
func (c *Config) GetPostgresConnectionMaxIdle() int {
	return c.v.GetInt(varPostgresConnectionMaxIdle)
}

// GetPostgresConnectionMaxOpen returns the max number of open connections that should be open in the database connection pool.
// -1 represents no restrictions/default behavior
func (c *Config) GetPostgresConnectionMaxOpen() int {
	return c.v.GetInt(varPostgresConnectionMaxOpen)
}

// GetPostgresConfigString returns a ready to use string for usage in sql.Open()
func (c *Config) GetPostgresConfigString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
		c.GetPostgresHost(),
		c.GetPostgresPort(),
		c.GetPostgresUser(),
		c.GetPostgresPassword(),
		c.GetPostgresDatabase(),
		c.GetPostgresSSLMode(),
		c.GetPostgresConnectionTimeout(),
	)
}

// GetAuthURL returns Auth service URL
func (c *Config) GetAuthURL() string {
	return c.v.GetString(varAuthURL)
}

// GetAuthGrantType returns the fabric8-auth Grant type used while retrieving
// user account token
func (c *Config) GetAuthGrantType() string {
	return "client_credentials"
}

// GetAuthClientID returns the tenant's client id used while
// communicating with fabric8-auth
func (c *Config) GetAuthClientID() string {
	return c.v.GetString(varAuthClientID)
}

// GetClientSecret returns the secret which will be used in
// conjunction with the tenant client id
func (c *Config) GetClientSecret() string {
	return c.v.GetString(varClientSecret)
}

// GetTokenKey returns the encryption key/passphrase which will be used
// to decrypt the cluster tokens stored in auth token mgm
func (c *Config) GetTokenKey() string {
	return c.v.GetString(varAuthTokenKey)
}

// IsCleanTestDataEnabled return true if we want to have clean data enabled
func (c *Config) IsCleanTestDataEnabled() bool {
	return c.v.GetBool(varCleanTestDataEnabled)
}

func (c *Config) GetSentryDSN() string {
	return c.v.GetString(varSentryDSN)
}

// Return True if we want to have DB Logs Enabled
func (c *Config) IsDBLogsEnabled() bool {
	return c.v.GetBool(varDBLogsEnabled)
}
