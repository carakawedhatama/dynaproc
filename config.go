package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	AppName     string
	AppVersion  string
	Environment string
	Database    DatabaseConfig
	RabbitMQ    RabbitMQConfig
	Dynamics365 Dynamics365Config
	GlitchTip   GlitchTipConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

type RabbitMQConfig struct {
	URL string
}

type Dynamics365Config struct {
	APIURL string
}

type GlitchTipConfig struct {
	APIURL string
}

func (c *Config) LoadConfig(path string) {
	viper.AddConfigPath(".")
	viper.SetConfigName(path)

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("fatal error config file: default \n", err)
		os.Exit(1)
	}

	for _, k := range viper.AllKeys() {
		value := viper.GetString(k)
		if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
			viper.Set(k, getEnvOrPanic(strings.TrimSuffix(strings.TrimPrefix(value, "${"), "}")))
		}
	}

	err = viper.Unmarshal(c)
	if err != nil {
		fmt.Println("fatal error config file: default \n", err)
		os.Exit(1)
	}
}

func getEnvOrPanic(env string) string {
	split := strings.Split(env, ":")
	res := os.Getenv(split[0])
	if len(res) == 0 {
		if len(split) > 1 {
			res = strings.Join(split[1:], ":")
		}
		if len(res) == 0 {
			panic("Mandatory env variable not found:" + env)
		}
	}
	return res
}
