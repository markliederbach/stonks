package config

import (
	"fmt"
	"os"

	"github.com/alpacahq/alpaca-trade-api-go/common"
	"github.com/sirupsen/logrus"
)

const (
	// AlpacaAPIBaseURLVariable specifies the base URL
	AlpacaAPIBaseURLVariable string = "APCA_API_BASE_URL"

	// LogLevelVariable specifies the logging level
	LogLevelVariable string = "APCA_LOG_LEVEL"

	// DefaultLogLevel specifies the default logging level
	DefaultLogLevel logrus.Level = logrus.InfoLevel
)

// Config holds all configuration data about the currently-running service
type Config struct {
	// Required variables
	AlpacaAPIBaseURL   string
	AlpacaAPIKeyID     string
	AlpacaAPISecretKey string

	// Optional variables
	LogLevel logrus.Level
}

// Load creates a new instance of Config, using all available
// defaults and overrides.
func Load() Config {
	config := Config{
		// Required
		AlpacaAPIBaseURL:   fromEnvString(AlpacaAPIBaseURLVariable, true, ""),
		AlpacaAPIKeyID:     fromEnvString(common.EnvApiKeyID, true, ""),
		AlpacaAPISecretKey: fromEnvString(common.EnvApiSecretKey, true, ""),

		// Optional
		LogLevel: fromEnvLogLevel(LogLevelVariable, false, DefaultLogLevel),
	}

	config.configureLogger()

	return config
}

func (c *Config) configureLogger() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)

	switch c.LogLevel {
	case logrus.TraceLevel:
		logrus.SetReportCaller(true)
	}

	logrus.SetLevel(c.LogLevel)
}

func fromEnvString(variable string, required bool, defaultValue string) string {
	rawValue, exists := fromEnv(variable, required)
	if !exists {
		rawValue = defaultValue
	}
	return rawValue
}

func fromEnvLogLevel(variable string, required bool, defaultValue logrus.Level) logrus.Level {
	var err error
	value := defaultValue
	rawValue, exists := fromEnv(variable, required)
	if exists {
		value, err = logrus.ParseLevel(rawValue)
		if err != nil {
			panic(err)
		}
	}
	return value
}

func fromEnv(variable string, required bool) (string, bool) {
	value, exists := os.LookupEnv(variable)
	if !exists && required {
		panic(fmt.Errorf("Missing required environment variable %s", variable))
	}
	return value, exists
}
