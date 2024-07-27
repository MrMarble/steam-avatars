package main

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mrmarble/steam-avatars/components"
	"github.com/patrickmn/go-cache"
)

type Context struct {
	SteamAPIKey  string
	client       *http.Client
	cache        *cache.Cache
	lastSearches []string
	m            sync.Mutex
	echo.Context
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
		templates: template.Must(template.New("hack.tmpl").Funcs(template.FuncMap{
			"url": func(s string) template.URL {
				return template.URL(s)
			},
		}).ParseGlob("public/templates/*")),
	}

	e.Renderer = t
	limiter := middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
		Rate:      20,
		Burst:     10,
		ExpiresIn: 1 * time.Minute,
	})
	rateLimiterConfig := middleware.DefaultRateLimiterConfig
	rateLimiterConfig.Store = limiter
	rateLimiterConfig.Skipper = func(c echo.Context) bool {
		// Skip rate limiter for requests with query param "key"
		return c.QueryParams().Has("key")
	}

	cache := cache.New(24*time.Hour, 1*time.Hour)
	cc := &Context{os.Getenv("STEAM_API_KEY"), &http.Client{}, cache, []string{}, sync.Mutex{}, nil}

	e.Group("", func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set(echo.HeaderCacheControl, "public, max-age=86400")
			return next(c)
		}
	}).Static("/static", "public")

	// Middleware
	e.Use(middleware.RateLimiterWithConfig(rateLimiterConfig))
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cacheKey := c.Request().URL.Path
			if !strings.HasPrefix(cacheKey, "/avatar/") {
				return next(c)
			}
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
			cc.Context = c
			return next(cc)
		}
	})
	e.Use(middleware.Logger())
	e.Use(middleware.Gzip())
	e.Use(middleware.CORS())
	e.Use(middleware.Secure())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", home)
	e.POST("/", search)
	e.GET("/avatar/:name", avatar)
	e.GET("/profile/:name", profile)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "1323"
	}
	e.Logger.Fatal(e.Start(":" + port))
}

func home(c echo.Context) error {

	return components.Index().Render(c.Request().Context(), c.Response().Writer)
}

func search(c echo.Context) error {
	name := c.FormValue("name")
	target := c.FormValue("target")
	cc := c.(*Context)

	if name == "" {
		return c.Render(http.StatusOK, "home", struct{ Error string }{"Please enter a Steam username or ID"})
	}
	steamid, err := steamID(name, c.(*Context))
	if err != nil {
		return c.Render(http.StatusOK, "home", struct{ Error string }{err.Error()})
	}

	found := false
	for _, id := range cc.lastSearches {
		if id == steamid {
			found = true
			break
		}
	}
	if !found {
		cc.m.Lock()
		if len(cc.lastSearches) >= 5 {
			cc.lastSearches = cc.lastSearches[1:]
		}
		cc.lastSearches = append(cc.lastSearches, steamid)
		cc.m.Unlock()
	}

	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s/%s", target, steamid))
}

func steamID(name string, cc *Context) (string, error) {
	if isSteamID(name) {
		return name, nil
	}

	if steamID, ok := cc.cache.Get(name); ok {
		return steamID.(string), nil
	}

	steamID, err := GetSteamID(cc.client, cc.SteamAPIKey, name)
	if err != nil {
		return "", err
	}

	cc.cache.Set(name, steamID, cache.DefaultExpiration)
	return steamID, nil
}

func downloadFile(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
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

	cc.cache.Set(c.Request().URL.Path, *data, cache.DefaultExpiration)
	if c.QueryParam("format") == "json" {
		return c.JSON(http.StatusOK, data)
	}
	c.Response().Header().Set(echo.HeaderContentType, "image/svg+xml")
	//return c.Render(http.StatusOK, "avatar", data)
	//return c.Blob(http.StatusOK, "image/svg+xml", []byte(data.Html))
	return components.Avatar(data.SteamID, data.AvatarURL, data.FrameURL).Render(c.Request().Context(), c.Response().Writer)
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
	isAnimated := true
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
		isAnimated = false
		avatar, err = GetAvatar(client, steamAPIKey, steamID)
		if err != nil {
			return nil, err
		}
	}
	avatarFile, err := downloadFile(avatar)
	if err != nil {
		return nil, err
	}
	if isAnimated {
		avatar = fmt.Sprintf("data:image/gif;base64,%s", base64.StdEncoding.EncodeToString(avatarFile))
	} else {
		avatar = fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(avatarFile))
	}

	frame, err := GetAvatarFrame(client, steamAPIKey, steamID)
	if err != nil {
		return nil, err
	}

	frameFile, err := downloadFile(frame)
	if err != nil {
		return nil, err
	}
	frame = fmt.Sprintf("data:image/apng;base64,%s", base64.StdEncoding.EncodeToString(frameFile))

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
