package server

import (
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

//go:generate templ generate "internal/server/templates/*"

func setupRoutes(e *echo.Echo) {
	e.Group("", func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Cache-Control", "public, max-age=604800")
			return next(c)
		}
	}).Static("/static", "assets")

	e.GET("/", handleIndex)
	e.POST("/", handleSearch)
	e.GET("/avatar/:steamID", handleAvatar)
}

func renderView(c echo.Context, cmp templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)

	return cmp.Render(c.Request().Context(), c.Response().Writer)
}

func renderSVG(c echo.Context, cmp templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, "image/svg+xml")

	return cmp.Render(c.Request().Context(), c.Response().Writer)
}
