package logger

import (
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config представляет конфигурацию логгера.
type Config struct {
	Level         string   `mapstructure:"level"`         // Уровень логирования
	Encoding      string   `mapstructure:"encoding"`      // Формат логирования (json, console)
	OutputPaths   []string `mapstructure:"output_paths"`   // Пути вывода логов
	ErrorOutputPaths []string `mapstructure:"error_output_paths"` // Пути вывода ошибок
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
		output, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("не удалось открыть файл для логирования: %s, %w", path, err)
		}
		core := zapcore.NewCore(encoder, zapcore.AddSync(output), level)
		cores = append(cores, core)
	}

	// Создаем core, который будет записывать ошибки в указанные error output paths
	errorCores := []zapcore.Core{}
	for _, path := range cfg.ErrorOutputPaths {
		output, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("не удалось открыть файл для ошибок логирования: %s, %w", path, err)
		}
		core := zapcore.NewCore(encoder, zapcore.AddSync(output), zap.ErrorLevel) // Логируем только ошибки и выше
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