package main

import (
	"fmt"
	"log"
	"os"

	"QiptFlow/internal/config"
	"QiptFlow/internal/logger"
	"QiptFlow/internal/server"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Init(false)
	if err != nil {
		log.Fatalf("Ошибка инициализации конфигурации: %v", err)
	}

	logCfg := logger.Config{
		Level:          cfg.LogLevel,
		Encoding:       cfg.Logger.Encoding,
		OutputPaths:    cfg.Logger.OutputPaths,
		ErrorOutputPaths: cfg.Logger.ErrorOutputPaths,
		LogRotation: cfg.Logger.LogRotation,
	}

	if err := logger.Init(logCfg); err != nil {
		log.Fatalf("Ошибка инициализации логгера: %v", err)
	}

	defer func() {
		if err := logger.GetLogger().Sync(); err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка при сбросе буфера логов: %v\n", err)
		}
	}()

	if err := server.Run(cfg); err != nil {
		logger.GetLogger().Fatal("Ошибка запуска сервера", zap.Error(err))
	}

	logger.GetLogger().Info("Приложение завершено")
}