package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// Config представляет структуру конфигурации приложения.
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

// loadConfig загружает конфигурацию из файла и переменных окружения.
func loadConfig(useEnv bool) error {
	// Инициализация валидатора
	validate = validator.New()

	viper.SetConfigType("yaml")
	viper.SetConfigFile("/app/configs/backend/config.yaml")

	// Основные значения по умолчанию
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("log_level", "info")

	// Проверка существования файла
	if _, err = os.Stat(viper.ConfigFileUsed()); os.IsNotExist(err) {
		return fmt.Errorf("конфигурационный файл не найден: %s", viper.ConfigFileUsed())
	}

	// Чтение конфигурации
	if err = viper.ReadInConfig(); err != nil {
		return fmt.Errorf("ошибка при чтении конфигурационного файла: %w", err)
	}

	// Использование переменных окружения
	if useEnv {
		viper.AutomaticEnv() // Переопределяет yaml
	}

	instance = &Config{}

	// Преобразование конфигурации в структуру
	if err = viper.Unmarshal(instance); err != nil {
		return fmt.Errorf("ошибка при преобразовании конфигурации: %w", err)
	}

	// Валидация значения конфигурации
	if err = validate.Struct(instance); err != nil {
		return fmt.Errorf("ошибка валидации конфигурации: %w", err)
	}

	// Вывод всех настроек (для отладки)
	allSettings := viper.AllSettings()
	fmt.Printf("Все настройки:\n%+v\n", allSettings)

	return nil
}

// Init инициализирует конфигурацию приложения.
func Init(configPath string, useEnv bool) (*Config, error) {
	once.Do(func() {
		err = loadConfig(useEnv)
	})

	if err != nil {
		// Возвращаем nil instance при ошибке
		instance = nil
	}

	return instance, err
}

// GetConfig возвращает экземпляр конфигурации.
// Возвращает ошибку, если конфигурация не была инициализирована.
func GetConfig() (*Config, error) {
	if instance == nil {
		return nil, fmt.Errorf("конфигурация не инициализирована, вызовите Init()")
	}
	return instance, nil
}