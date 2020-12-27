package config_test

import (
	"os"
	"strings"

	"github.com/alpacahq/alpaca-trade-api-go/common"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"

	"github.com/markliederbach/stonks/pkg/alpaca/config"
)

var _ = Describe("Config", func() {
	Describe("Loading config from the environment", func() {
		var (
			appConfig            config.Config
			preservedEnvironment map[string]string
			variablePrefix       = "APCA_"
			baseURL              = "https://example.url"
			keyID                = "mykey"
			secretKey            = "secret123"
		)

		BeforeEach(func() {
			preservedEnvironment = preserveAndClearEnvironment(variablePrefix)
		})

		AfterEach(func() {
			resetEnvironment(variablePrefix, preservedEnvironment)
		})

		Context("when only required variables are present", func() {
			BeforeEach(func() {
				os.Setenv(config.AlpacaAPIBaseURLVariable, baseURL)
				os.Setenv(common.EnvApiKeyID, keyID)
				os.Setenv(common.EnvApiSecretKey, secretKey)

				appConfig = config.Load()
			})

			It("should set the required variables on the config object", func() {
				Expect(appConfig.AlpacaAPIBaseURL).To(Equal(baseURL))
				Expect(appConfig.AlpacaAPIKeyID).To(Equal(keyID))
				Expect(appConfig.AlpacaAPISecretKey).To(Equal(secretKey))
			})

			It("should set default optional variables on the config object", func() {
				Expect(appConfig.LogLevel).To(Equal(config.DefaultLogLevel))
			})
		})

		Context("when optional variables are set", func() {
			BeforeEach(func() {
				os.Setenv(config.AlpacaAPIBaseURLVariable, baseURL)
				os.Setenv(common.EnvApiKeyID, keyID)
				os.Setenv(common.EnvApiSecretKey, secretKey)

				os.Setenv(config.LogLevelVariable, "DEBUG")

				appConfig = config.Load()
			})

			It("should set optional variables on the config object", func() {
				Expect(appConfig.LogLevel).To(Equal(logrus.DebugLevel))
			})
		})

		Context("when log level is not parsable", func() {
			BeforeEach(func() {
				os.Setenv(config.AlpacaAPIBaseURLVariable, baseURL)
				os.Setenv(common.EnvApiKeyID, keyID)
				os.Setenv(common.EnvApiSecretKey, secretKey)

				os.Setenv(config.LogLevelVariable, "FOOBAR")

			})

			It("should panic", func() {
				Expect(func() { config.Load() }).To(Panic())
			})
		})

		Context("when required Base URL is not set", func() {
			BeforeEach(func() {
				os.Setenv(common.EnvApiKeyID, keyID)
				os.Setenv(common.EnvApiSecretKey, secretKey)
			})

			It("should panic", func() {
				Expect(func() { config.Load() }).To(Panic())
			})
		})

		Context("when required Key ID is not set", func() {
			BeforeEach(func() {
				os.Setenv(config.AlpacaAPIBaseURLVariable, baseURL)
				os.Setenv(common.EnvApiSecretKey, secretKey)
			})

			It("should panic", func() {
				Expect(func() { config.Load() }).To(Panic())
			})
		})

		Context("when required Secret Key is not set", func() {
			BeforeEach(func() {
				os.Setenv(config.AlpacaAPIBaseURLVariable, baseURL)
				os.Setenv(common.EnvApiKeyID, keyID)
			})

			It("should panic", func() {
				Expect(func() { config.Load() }).To(Panic())
			})
		})
	})
})

// preserveAndClearEnvironment stores and clears existing values into a map to be restored later
func preserveAndClearEnvironment(prefix string) map[string]string {
	var preservedEnvironment = map[string]string{}
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, prefix) {
			continue
		}

		parts := strings.SplitN(env, "=", 2)
		preservedEnvironment[parts[0]] = parts[1]

		os.Unsetenv(parts[0])
	}
	return preservedEnvironment
}

// resetEnvironment resets any preserved variables to their original values
func resetEnvironment(prefix string, preservedEnvironment map[string]string) {
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, prefix) {
			continue
		}

		parts := strings.SplitN(env, "=", 2)
		restoredValue, exists := preservedEnvironment[parts[0]]

		if !exists {
			os.Unsetenv(parts[0])
			delete(preservedEnvironment, parts[0])
			continue
		}

		if restoredValue != parts[1] {
			os.Setenv(parts[0], restoredValue)
		}

	}
}
