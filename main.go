package main

import (
	"fmt"
	"html/template"
	"io"
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

type AvatarResponse struct {
	SteamID   string `json:"steamid"`
	AvatarURL string `json:"avatar_url"`
	FrameURL  string `json:"frame_url"`
	Html      string `json:"html"`
}

type ProfileResponse struct {
	AvatarResponse
	WebmURL  string `json:"webm_url"`
	Mp4URL   string `json:"mp4_url"`
	ImageURL string `json:"image_url"`
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	// Echo instance
	e := echo.New()
	t := &Template{
		templates: template.Must(template.ParseGlob("templates/*.svg")),
	}
	e.Renderer = t
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
					switch data := cached.(type) {
					case AvatarResponse:
						c.Response().Header().Set(echo.HeaderContentType, "image/svg+xml")
						return c.Render(http.StatusOK, "avatar", data)
					case ProfileResponse:
						c.Response().Header().Set(echo.HeaderContentType, "image/svg+xml")
						return c.Render(http.StatusOK, "profile", data)
					}
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
	e.GET("/profile/:name", profile)

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

	data, err := fetchAvatarData(name, cc.client, steamAPIKey)
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
	c.Response().Header().Set(echo.HeaderContentType, "image/svg+xml")
	return c.Render(http.StatusOK, "avatar", data)
	//return c.Blob(http.StatusOK, "image/svg+xml", []byte(data.Html))
}

// Handler
func profile(c echo.Context) error {
	name := c.Param("name")
	cc := c.(*Context)

	steamAPIKey := cc.SteamAPIKey
	if c.QueryParams().Has("key") {
		steamAPIKey = c.QueryParam("key")
	}

	data, err := fetchProfileData(name, cc.client, steamAPIKey)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	data.Html = fmt.Sprintf(`<svg width="640" height="570" viewbox="0 0 640 570" xmlns="http://www.w3.org/2000/svg">
	<title>Steam avatar of %s</title>
	<desc>Generated with https://github.com/mrmarble/steam-avatars</desc>
	<g>
		<foreignObject width="640" height="570">
			<video xmlns="http://www.w3.org/1999/xhtml" poster="%s" width="640" height="570" autoplay="" loop="">
			<source src="%s" type="video/webm"/>
			<source src="%s" type="video/mp4"/>
				<source src="%[1]s"/>
			</video>
		</foreignObject>
		<g>
		<image id="svg_3" href="%s" height="184" width="184" y="40" x="40"/>
		<image id="svg_2" href="%s" height="224" width="224" y="20" x="20"/>
		</g>
	</g>
 </svg>`, data.SteamID, AssetURL+data.ImageURL, AssetURL+data.WebmURL, AssetURL+data.Mp4URL, data.AvatarURL, data.FrameURL)

	cc.cache.Set(c.Request().URL.Path, *data, cache.DefaultExpiration)
	if c.QueryParam("format") == "json" {
		return c.JSON(http.StatusOK, data)
	}
	c.Response().Header().Set(echo.HeaderContentType, "image/svg+xml")
	return c.Render(http.StatusOK, "profile", data)
	//return c.Blob(http.StatusOK, "image/svg+xml", []byte(data.Html))
}

func fetchAvatarData(name string, client *http.Client, steamAPIKey string) (*AvatarResponse, error) {
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

	return &AvatarResponse{
		SteamID:   steamID,
		AvatarURL: avatar,
		FrameURL:  frame,
	}, nil
}

func fetchProfileData(name string, client *http.Client, steamAPIKey string) (*ProfileResponse, error) {
	data, err := fetchAvatarData(name, client, steamAPIKey)
	if err != nil {
		return nil, err
	}

	background, err := GetMiniProfileBackground(client, steamAPIKey, data.SteamID)
	if err != nil {
		return nil, err
	}

	return &ProfileResponse{
		AvatarResponse: *data,
		WebmURL:        background.Response.ProfileBackground.MovieWebm,
		Mp4URL:         background.Response.ProfileBackground.MovieMP4,
		ImageURL:       background.Response.ProfileBackground.ImageLarge,
	}, nil
}
