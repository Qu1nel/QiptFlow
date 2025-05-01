package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Host string `mapstructure:"host" validate:"required,hostname_rfc1123|ip4_addr|eq=localhost"`
		Port int    `mapstructure:"port" validate:"required,min=1,max=65535"`
	} `mapstructure:"server"`
	LogLevel string `mapstructure:"log_level" validate:"omitempty,oneof=debug info warn error fatal"`
}

var (
	once sync.Once
	instance *Config
	err error
	validate *validator.Validate
)

func Init(configPath string, useEnv bool) (*Config, error) {
	once.Do(func() {
		// Валидатора
		validate = validator.New()

		viper.SetConfigType("yaml")
		viper.SetConfigFile("/app/configs/backend/config.yaml")

		// Основные значения по умолчанию
		viper.SetDefault("server.host", "localhost")
		viper.SetDefault("server.port", 8080)
		viper.SetDefault("log_level", "info")

		// Проверка существования файла
		if _, err = os.Stat(viper.ConfigFileUsed()); os.IsNotExist(err) {
			err = fmt.Errorf("конфигурационный файл не найден: %s", viper.ConfigFileUsed())
			return
		}

		// Чтение конфигурации
		if err = viper.ReadInConfig(); err != nil {
			err = fmt.Errorf("ошибка при чтении конфигурационного файла: %s", err)
			return
		}

		// Использование переменных окружения
		if useEnv {
			viper.AutomaticEnv()  // Переопределяет yaml
		}

		instance = &Config{}

		// Преобразование конфигурации в структуру
		if err = viper.Unmarshal(instance); err != nil {
			err = fmt.Errorf("ошибка при преобразовании конфигурации: %w", err)
			return
		}

		// Валидация значения конфигурации
		if err = validate.Struct(instance); err != nil {
			err = fmt.Errorf("ошибка валидации конфигурации: %w", err)
			return
		}

		// Вывод всех настроек (для отладки)
		allSettings := viper.AllSettings()
		fmt.Printf("Все настройки:\n%+v\n", allSettings)

	})

	return instance, err
}

func GetConfig() *Config {
	return instance
}