package config

import (
	"sync"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Host string `mapstructure:"host"`
		Port int `mapstructure:"port"`
	} `mapstructure:"server"`

	LogLever string `mapstructure:"log_level"`
}

var (
	once sync.Once
	instance *Config
)

func Init(configPath string, useEnv bool) (*Config, error) {
	var err error
	once.Do(func() {
		viper.SetConfigType("yaml")
		viper.SetConfigFile("/app/configs/backend/config.yaml")

		if useEnv {
			viper.AutomaticEnv()  // Переопределяет yaml
		}

		if err	= viper.ReadInConfig(); err != nil {
			return
		}

		instance = &Config{}
		if err = viper.Unmarshal(instance); err != nil {
			return
		}

	})
	return instance, err
}

func GetConfig() *Config {
	return instance
}