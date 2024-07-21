package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/patrickmn/go-cache"
)

type Context struct {
	echo.Context
	SteamAPIKey string
	client      *http.Client
	cache       *cache.Cache
}

type Response struct {
	SteamID   string `json:"steamid"`
	AvatarURL string `json:"avatar_url"`
	FrameURL  string `json:"frame_url"`
	Html      string `json:"html"`
}

func main() {
	// Echo instance
	e := echo.New()

	limiter := middleware.NewRateLimiterMemoryStore(1)
	rateLimiterConfig := middleware.DefaultRateLimiterConfig
	rateLimiterConfig.Store = limiter
	rateLimiterConfig.Skipper = func(c echo.Context) bool {
		// Skip rate limiter for requests with query param "key"
		return c.QueryParams().Has("key")
	}

	cache := cache.New(24*time.Hour, 1*time.Hour)

	// Middleware
	e.Use(middleware.RateLimiterWithConfig(rateLimiterConfig))
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cacheKey := c.Request().URL.Path

			// Set Cache-Control header
			c.Response().Header().Set(echo.HeaderCacheControl, "public, max-age=86400")

			if cached, found := cache.Get(cacheKey); found {
				if c.QueryParam("format") == "json" {
					return c.JSON(http.StatusOK, cached)
				} else {
					return c.Blob(http.StatusOK, "image/svg+xml", []byte(cached.(Response).Html))
				}
			}

			return next(c)
		}
	})
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &Context{c, os.Getenv("STEAM_API_KEY"), &http.Client{}, cache}
			return next(cc)
		}
	})
	e.Use(middleware.Logger())
	e.Use(middleware.Gzip())
	e.Use(middleware.CORS())
	e.Use(middleware.Secure())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/avatar/:name", avatar)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "1323"
	}
	e.Logger.Fatal(e.Start(":" + port))
}

// Handler
func avatar(c echo.Context) error {
	name := c.Param("name")
	cc := c.(*Context)

	steamAPIKey := cc.SteamAPIKey
	if c.QueryParams().Has("key") {
		steamAPIKey = c.QueryParam("key")
	}

	data, err := fetchSteamData(name, cc.client, steamAPIKey)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	data.Html = fmt.Sprintf(`<svg width="224" height="224" viewbox="0 0 224 224" xmlns="http://www.w3.org/2000/svg">
	<title>Steam avatar of %s</title>
	<desc>Generated with https://github.com/mrmarble/steam-avatars</desc>
	<g>
	 <image id="svg_3" href="%s" height="184" width="184" y="20" x="20"/>
	 <image id="svg_2" href="%s" height="224" width="224" y="0" x="0"/>
	</g>
 </svg>`, data.SteamID, data.AvatarURL, data.FrameURL)

	cc.cache.Set(c.Request().URL.Path, *data, cache.DefaultExpiration)
	if c.QueryParam("format") == "json" {
		return c.JSON(http.StatusOK, data)
	}
	return c.Blob(http.StatusOK, "image/svg+xml", []byte(data.Html))
}

func fetchSteamData(name string, client *http.Client, steamAPIKey string) (*Response, error) {
	// Check if name is a steamID
	steamID := name
	if !isSteamID(steamID) {
		var err error
		steamID, err = GetSteamID(client, steamAPIKey, name)
		if err != nil {
			return nil, err
		}
	}

	avatar, err := GetAnimatedAvatar(client, steamAPIKey, steamID)
	if err != nil {
		return nil, err
	}
	if avatar == "" {
		avatar, err = GetAvatar(client, steamAPIKey, steamID)
		if err != nil {
			return nil, err
		}
	}

	frame, err := GetAvatarFrame(client, steamAPIKey, steamID)
	if err != nil {
		return nil, err
	}

	return &Response{
		SteamID:   steamID,
		AvatarURL: avatar,
		FrameURL:  frame,
	}, nil
}
