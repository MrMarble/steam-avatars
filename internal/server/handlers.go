package server

import (
	"github.com/labstack/echo/v4"
	"github.com/mrmarble/steam-avatars/internal/server/templates"
	"github.com/mrmarble/steam-avatars/internal/steam"
)

func handleIndex(c echo.Context) error {
	page := templates.Index()
	return renderView(c, page)
}

func handleSearch(c echo.Context) error {
	cc := c.(*Context)
	name := c.FormValue("name")
	if name == "" {
		return c.JSON(400, map[string]string{"error": "name is required"})
	}

	c.Logger().Info("searching for vanity URL ", name)
	steamID := name

	if !steam.IsSteamID(name) {
		var err error
		steamID, err = cc.client.GetSteamID(name)
		if err != nil {
			return err
		}
	}

	frame, err := cc.client.GetAvatarFrame(steamID)
	if err != nil {
		return err
	}
	avatar, err := cc.client.GetAnimatedAvatar(steamID)
	if err != nil {
		return err
	}

	avatarTempl := templates.Result(steamID, avatar, frame, c.Request().URL.Scheme+"://"+c.Request().Host+"/avatar/"+steamID)

	return renderView(c, avatarTempl)

}

func handleAvatar(c echo.Context) error {
	cc := c.(*Context)
	steamID := c.Param("steamID")
	if !steam.IsSteamID(steamID) {
		var err error
		steamID, err = cc.client.GetSteamID(steamID)
		if err != nil {
			return err
		}
	}

	frame, err := cc.client.GetAvatarFrame(steamID)
	if err != nil {
		return err
	}
	avatar, err := cc.client.GetAnimatedAvatar(steamID)
	if err != nil {
		return err
	}

	return c.JSON(200, map[string]string{"avatar": avatar, "frame": frame})
}
