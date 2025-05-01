package config

import (
	"os"
	"path/filepath"
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
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")

		executablePath, err := os.Executable()
		if err != nil {
			panic(err)
		}
		executableDir := filepath.Dir(executablePath)
		configDir := filepath.Join(executableDir, "configs", "backend")

		viper.AddConfigPath(configDir)

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