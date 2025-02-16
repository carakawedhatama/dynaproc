package main

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func resetViper() {
	viper.Reset()
	viper.AddConfigPath(".")
}

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file for testing
	configContent := `
appName: "TestApp"
appVersion: "1.0.0"
environment: "test"
database:
  host: "localhost"
  port: 5432
  user: "${DB_USER:default_user}"
  password: "${DB_PASSWORD:default_pass}"
  name: "testdb"
  sslMode: "disable"
rabbitmq:
  url: "${RABBITMQ_URL:amqp://guest:guest@localhost:5672/}"
dynamics365:
  apiUrl: "${DYNAMICS_API_URL:https://api.example.com}"
glitchTip:
  apiUrl: "${GLITCHTIP_URL:https://glitchtip.example.com}"
`
	err := os.WriteFile("config_test.yaml", []byte(configContent), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("config_test.yaml")

	// Test case 1: Load config with environment variables
	t.Run("Load config with environment variables", func(t *testing.T) {
		resetViper()

		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("RABBITMQ_URL")

		os.Setenv("DB_USER", "test_user")
		os.Setenv("DB_PASSWORD", "test_pass")
		os.Setenv("RABBITMQ_URL", "amqp://test:test@localhost:5672/")

		config := &Config{}
		config.LoadConfig("config_test")

		assert.Equal(t, "TestApp", config.AppName)
		assert.Equal(t, "1.0.0", config.AppVersion)
		assert.Equal(t, "test", config.Environment)
		assert.Equal(t, "test_user", config.Database.User)
		assert.Equal(t, "test_pass", config.Database.Password)
		assert.Equal(t, "amqp://test:test@localhost:5672/", config.RabbitMQ.URL)

		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("RABBITMQ_URL")
	})

	// Test case 2: Load config with default values
	t.Run("Load config with default values", func(t *testing.T) {
		resetViper()

		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("RABBITMQ_URL")

		config := &Config{}
		config.LoadConfig("config_test")

		assert.Equal(t, "default_user", config.Database.User)
		assert.Equal(t, "default_pass", config.Database.Password)
		assert.Equal(t, "amqp://guest:guest@localhost:5672/", config.RabbitMQ.URL)
		assert.Equal(t, "https://api.example.com", config.Dynamics365.APIURL)
		assert.Equal(t, "https://glitchtip.example.com", config.GlitchTip.APIURL)
	})
}

func TestGetEnvOrPanic(t *testing.T) {
	// Test case 1: Environment variable exists
	t.Run("Environment variable exists", func(t *testing.T) {
		os.Setenv("TEST_VAR", "test_value")
		defer os.Unsetenv("TEST_VAR")

		result := getEnvOrPanic("TEST_VAR")
		assert.Equal(t, "test_value", result)
	})

	// Test case 2: Environment variable doesn't exist but has default value
	t.Run("Environment variable with default value", func(t *testing.T) {
		result := getEnvOrPanic("NONEXISTENT_VAR:default_value")
		assert.Equal(t, "default_value", result)
	})

	// Test case 3: Environment variable doesn't exist and no default value
	t.Run("Environment variable without default value should panic", func(t *testing.T) {
		assert.Panics(t, func() {
			getEnvOrPanic("NONEXISTENT_VAR")
		})
	})
}
