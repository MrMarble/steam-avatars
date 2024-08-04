package server

import (
	"context"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mrmarble/steam-avatars/internal/database"
	"github.com/mrmarble/steam-avatars/internal/steam"
	"github.com/rs/zerolog"
	"github.com/ziflex/lecho/v3"
)

type Server struct {
	e *echo.Echo
}

type Context struct {
	db     *database.Database
	client *steam.Client
	echo.Context
}

func NewServer(logger zerolog.Logger, db *database.Database, steamApiKey string) *Server {
	l := lecho.From(logger)
	e := echo.New()
	client := steam.NewClient(steamApiKey)

	e.HideBanner = true
	e.Logger = l

	e.Use(
		middleware.RequestID(),
		lecho.Middleware(lecho.Config{
			Logger:              l,
			NestKey:             "request",
			RequestLatencyLevel: zerolog.WarnLevel,
			RequestLatencyLimit: 1 * time.Second,
			Enricher: func(c echo.Context, logger zerolog.Context) zerolog.Context {
				return logger.Str("connecting-ip", c.Request().Header.Get("CF-Connecting-IP")).
					Str("country", c.Request().Header.Get("CF-IPCountry"))
			},
		}),
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				cc := &Context{db, client, c}
				return next(cc)
			}
		},
		middleware.Gzip(),
		middleware.CORS(),
		middleware.Secure(),
		middleware.Recover(),
	)

	setupRoutes(e)

	return &Server{e}
}

func (s *Server) Start() error {
	return s.e.Start(":8080")
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.e.Shutdown(ctx)
}
