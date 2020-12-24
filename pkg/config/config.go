package config

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

const (
	alpacaAPIBaseURLVariable   string = "ALPACA_API_BASE_URL"
	alpacaAPIKeyIDVariable     string = "ALPACA_API_KEY_ID"
	alpacaAPISecretKeyVariable string = "ALPACA_API_SECRET_KEY"
	logLevelVariable           string = "LOG_LEVEL"

	defaultLogLevel logrus.Level = logrus.InfoLevel
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
		AlpacaAPIBaseURL:   fromEnvString(alpacaAPIBaseURLVariable, true, ""),
		AlpacaAPIKeyID:     fromEnvString(alpacaAPIKeyIDVariable, true, ""),
		AlpacaAPISecretKey: fromEnvString(alpacaAPISecretKeyVariable, true, ""),

		// Optional
		LogLevel: fromEnvLogLevel(logLevelVariable, false, defaultLogLevel),
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
