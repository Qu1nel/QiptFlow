package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	server := echo.New()
	server.Debug = true
	server.Use(middleware.Logger())

	server.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "QiptFlow is running!")
	})

	server.Logger.Fatal(server.Start(":8080"))
}
