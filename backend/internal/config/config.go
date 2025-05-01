package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"

	"QiptFlow/internal/logger"
)

// Config представляет структуру конфигурации приложения.
type Config struct {
	Server struct {
		Host string `mapstructure:"host" validate:"required,hostname_rfc1123|ip4_addr|eq=localhost"`
		Port int    `mapstructure:"port" validate:"required,min=1,max=65535"`
	} `mapstructure:"server"`
	LogLevel string `mapstructure:"log_level" validate:"omitempty,oneof=debug info warn error fatal"`
	CORS     struct {
		AllowOrigins []string `mapstructure:"allow_origins"`
	} `mapstructure:"cors"`
	Gzip struct {
		Level int `mapstructure:"level"`
	} `mapstructure:"gzip"`
	Logger loggerConfig `mapstructure:"logger"`
}

type loggerConfig struct {
	Encoding      string   `mapstructure:"encoding" validate:"required,oneof=json console"`
	OutputPaths   []string `mapstructure:"output_paths" validate:"required,min=1"`
	ErrorOutputPaths []string `mapstructure:"error_output_paths" validate:"required,min=1"`
	LogRotation logger.LogRotationConfig `mapstructure:"log_rotation"`
}

var (
	once     sync.Once
	instance *Config
	err      error
	validate *validator.Validate
)

// loadConfig загружает конфигурацию из файла и переменных окружения.
func loadConfig(useEnv bool) error {
	validate = validator.New()

	viper.SetConfigType("yaml")
	viper.SetConfigFile("/app/configs/backend/config.yaml")

	// Основные значения по умолчанию
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("log_level", "info")
	viper.SetDefault("cors.allow_origins", []string{"*"})
	viper.SetDefault("gzip.level", 5)
	viper.SetDefault("logger.encoding", "json") // Значение по умолчанию для encoding
	viper.SetDefault("logger.output_paths", []string{"stdout"}) // Значение по умолчанию для output_paths
	viper.SetDefault("logger.error_output_paths", []string{"stderr"}) // Значение по умолчанию для error_output_paths
	viper.SetDefault("logger.log_rotation.enabled", false)
	viper.SetDefault("logger.log_rotation.filename", "application.log")
	viper.SetDefault("logger.log_rotation.max_size", 100)
	viper.SetDefault("logger.log_rotation.max_backups", 5)
	viper.SetDefault("logger.log_rotation.max_age", 30)
	viper.SetDefault("logger.log_rotation.compress", true)

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

	return nil
}

// Init инициализирует конфигурацию приложения.
func Init(useEnv bool) (*Config, error) {
	once.Do(func() {
		err = loadConfig(useEnv)
	})

	if err != nil {
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