package logger

import (
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Config представляет конфигурацию логгера.
type Config struct {
	Level         string   `mapstructure:"level"`         // Уровень логирования
	Encoding      string   `mapstructure:"encoding"`      // Формат логирования (json, console)
	OutputPaths   []string `mapstructure:"output_paths"`   // Пути вывода логов
	ErrorOutputPaths []string `mapstructure:"error_output_paths"` // Пути вывода ошибок
	LogRotation   LogRotationConfig `mapstructure:"log_rotation"` // Конфигурация ротации логов
}

// LogRotationConfig представляет конфигурацию ротации логов.
type LogRotationConfig struct {
	Enabled    bool   `mapstructure:"enabled"`    // Включена ли ротация логов
	Filename   string `mapstructure:"filename"`   // Имя файла для ротации
	MaxSize    int    `mapstructure:"max_size"`    // Максимальный размер файла в мегабайтах
	MaxBackups int    `mapstructure:"max_backups"` // Максимальное количество старых файлов логов
	MaxAge     int    `mapstructure:"max_age"`     // Максимальный срок хранения старых файлов логов в днях
	Compress   bool   `mapstructure:"compress"`   // Сжимать ли старые файлы логов
}

var (
	log      *zap.Logger
	once     sync.Once
	err      error
)

// Init инициализирует логгер.
func Init(cfg Config) error {
	once.Do(func() {
		err = loadLogger(cfg)
	})
	return err
}

// loadLogger загружает и конфигурирует логгер
func loadLogger(cfg Config) error {
	level := zap.NewAtomicLevel()
	if err = level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return fmt.Errorf("ошибка при разборе уровня логирования: %w", err)
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:    zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var encoder zapcore.Encoder
	switch cfg.Encoding {
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	case "console":
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		return fmt.Errorf("неподдерживаемый формат логирования: %s", cfg.Encoding)
	}

	// Создаем core, который будет записывать логи в указанные output paths
	cores := []zapcore.Core{}
	for _, path := range cfg.OutputPaths {
		var writeSyncer zapcore.WriteSyncer
		switch path {
		case "stdout":
			writeSyncer = os.Stdout
		case "stderr":
			writeSyncer = os.Stderr
		default:
			// Если указан другой путь, используем lumberjack для ротации логов
			if cfg.LogRotation.Enabled {
				lumberJackLogger := &lumberjack.Logger{
					Filename:   cfg.LogRotation.Filename,
					MaxSize:    cfg.LogRotation.MaxSize,    // megabytes
					MaxBackups: cfg.LogRotation.MaxBackups, // Максимальное количество старых файлов логов
					MaxAge:     cfg.LogRotation.MaxAge,     // days
					Compress:   cfg.LogRotation.Compress,   // Следует ли сжимать архивные файлы с помощью gzip
				}
				writeSyncer = zapcore.AddSync(lumberJackLogger)
			} else {
				output, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
				if err != nil {
					return fmt.Errorf("не удалось открыть файл для логирования: %s, %w", path, err)
				}
				writeSyncer = zapcore.AddSync(output)
			}
		}
		core := zapcore.NewCore(encoder, writeSyncer, level)
		cores = append(cores, core)
	}

	// Создаем core, который будет записывать ошибки в указанные error output paths
	errorCores := []zapcore.Core{}
	for _, path := range cfg.ErrorOutputPaths {
		var writeSyncer zapcore.WriteSyncer
		switch path {
		case "stdout":
			writeSyncer = os.Stdout
		case "stderr":
			writeSyncer = os.Stderr
		default:
			output, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
			if err != nil {
				return fmt.Errorf("не удалось открыть файл для ошибок логирования: %s, %w", path, err)
			}
			writeSyncer = zapcore.AddSync(output)
		}
		core := zapcore.NewCore(encoder, writeSyncer, zap.ErrorLevel) // Логируем только ошибки и выше
		errorCores = append(errorCores, core)
	}

	// Объединяем все core
	allCores := append(cores, errorCores...)
	combinedCore := zapcore.NewTee(allCores...)

	// Добавляем CallerSkip, чтобы в логах отображался вызывающий код, а не код логгера
	log = zap.New(combinedCore,
		zap.AddCaller(), // Добавляем информацию о вызывающем коде
		zap.AddCallerSkip(1), // Пропускаем один уровень вызова, чтобы видеть вызывающий код, а не код логгера
		zap.AddStacktrace(zap.ErrorLevel), // Добавляем stacktrace для ошибок и выше
	)

	return nil
}

// GetLogger возвращает экземпляр логгера.
func GetLogger() *zap.Logger {
	if log == nil {
		fmt.Fprintf(os.Stderr, "Logger не был инициализирован. Используйте Init() для инициализации логгера.\n")
		os.Exit(1)
	}
	return log
}