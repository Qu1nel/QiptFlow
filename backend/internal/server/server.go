package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"QiptFlow/internal/config"
	"QiptFlow/internal/logger"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

// Run запускает HTTP-сервер.
func Run(cfg *config.Config) error {
	e := echo.New()
	e.HideBanner = true
	configureMiddleware(e, cfg)
	registerRoutes(e)
	return startServer(e, cfg)
}

// configureMiddleware настраивает middleware для Echo.
func configureMiddleware(e *echo.Echo, cfg *config.Config) {
	log := logger.GetLogger()

	e.Use(middleware.RequestID()) // Request ID middleware для отслеживания запросов.
	e.Use(middleware.Recover()) // Recover middleware для обработки паник.

	// CORS middleware для разрешения запросов с других доменов.
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: cfg.CORS.AllowOrigins,
	}))

	// Gzip middleware для сжатия ответов.
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: cfg.Gzip.Level,
	}))

	// Request Logger middleware для логирования запросов.
	e.Use(requestLoggerMiddleware(log))
}

// registerRoutes регистрирует маршруты для Echo.
func registerRoutes(e *echo.Echo) {
	e.GET("/health", healthCheckHandler)
	e.GET("/", helloHandler)
}

// startServer запускает HTTP-сервер с graceful shutdown.
func startServer(e *echo.Echo, cfg *config.Config) error {
	log := logger.GetLogger()

	go func() {
		address := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
		log.Info("Запуск HTTP-сервера", zap.String("адрес", address))
		if err := e.Start(address); err != nil && err != http.ErrServerClosed {
			log.Fatal("Ошибка запуска HTTP-сервера", zap.Error(err))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Info("Выключение HTTP-сервера")
	if err := e.Shutdown(ctx); err != nil {
		log.Error("Ошибка выключения HTTP-сервера", zap.Error(err))
		return err
	}

	log.Info("HTTP-сервер выключен")
	return nil
}

// requestLoggerMiddleware создает middleware для логирования запросов.
func requestLoggerMiddleware(log *zap.Logger) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			log.Info("Запрос",
				zap.String("uri", v.URI),
				zap.Int("status", v.Status),
				zap.String("method", c.Request().Method),
				zap.String("remote_addr", c.RealIP()),
				zap.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)),
			)
			return nil
		},
	})
}

// healthCheckHandler обрабатывает запросы health check.
func healthCheckHandler(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

// helloHandler обрабатывает запросы на корневой роут.
func helloHandler(c echo.Context) error {
	return c.String(http.StatusOK, "QiptFlow is running!")
}
