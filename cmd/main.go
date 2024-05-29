package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	//TODO: see alternatives of logging
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		LogProtocol: true,
		HandleError: true, // forwards error to the global error handler, so it can decide appropriate status code
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				// logger.LogAttrs(context.Background(), slog.LevelInfo, "REQUEST",
				// 	slog.String("uri", v.URI),
				// 	slog.Int("status", v.Status),
				// )

				logger.InfoContext(c.Request().Context(), "", slog.Group("Request",
					slog.String("uri", v.URI),
					slog.String("method", v.Method),
					slog.Int("status", v.Status),
					slog.String("Protocol", v.Protocol),
				))
			} else {
				logger.LogAttrs(context.Background(), slog.LevelError, "REQUEST_ERROR",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("err", v.Error.Error()),
				)
			}
			return nil
		},
	}))

	e.GET("/", func(c echo.Context) error {
		return c.String(200, "hello")
	})
	e.Logger.Fatal(e.Start(":3000"))
}
