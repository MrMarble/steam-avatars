package server

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/mrmarble/steam-avatars/internal/database"
	"github.com/mrmarble/steam-avatars/internal/server/templates"
	"github.com/mrmarble/steam-avatars/internal/steam"
	"github.com/valkey-io/valkey-go"
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

	var err error
	var user *database.User
	if steam.IsSteamID(name) {
		id, _ := strconv.ParseInt(name, 10, 64)
		user, err = cc.db.GetUserByID(id)
	} else {
		user, err = cc.db.GetUserByVanityURL(name)
	}
	if err != nil && !valkey.IsValkeyNil(err) {
		return fmt.Errorf("failed to search for user: %w", err)
	}

	if user == nil {
		user, err = searchUser(cc.client, name)
		if err != nil {
			return err
		}
		err = cc.db.CreateUser(user)
		if err != nil {
			return err
		}
	}

	strID := strconv.FormatInt(user.ID, 10)
	return renderView(c, templates.Result(strID, user.Avatar, user.Frame, c.Request().URL.Scheme+"://"+c.Request().Host+"/avatar/"+strID))
}

func searchUser(c *steam.Client, query string) (*database.User, error) {
	steamID, err := c.GetSteamID(query)
	if err != nil {
		return nil, err
	}

	summary, err := c.GetPlayer(steamID)
	if err != nil {
		return nil, err
	}

	frame, err := c.GetAvatarFrame(steamID)
	if err != nil {
		return nil, err
	}
	/*frameFile, err := downloadFile(frame)
	if err != nil {
		return nil, err
	}
	frame = fmt.Sprintf("data:image/apng;base64,%s", base64.StdEncoding.EncodeToString(frameFile))
	*/
	avatar, err := c.GetAnimatedAvatar(steamID)
	if err != nil {
		return nil, err
	}

	/*	if avatar == "" {
			avatarFile, err := downloadFile(summary.AvatarFull)
			if err != nil {
				return nil, err
			}
			avatar = fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(avatarFile))
		} else {
			avatarFile, err := downloadFile(avatar)
			if err != nil {
				return nil, err
			}
			avatar = fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(avatarFile))
		}*/

	ID, _ := strconv.ParseInt(steamID, 10, 64)

	if avatar == "" {
		avatar = summary.AvatarFull
	}
	return &database.User{
		ID:          ID,
		VanityURL:   query,
		DisplayName: summary.PersonaName,
		Avatar:      avatar,
		Frame:       frame,
	}, nil
}

func handleAvatar(c echo.Context) error {
	cc := c.(*Context)
	steamID := c.Param("steamID")
	if !steam.IsSteamID(steamID) {
		return c.JSON(400, map[string]string{"error": "invalid steamID"})
	}

	ID, _ := strconv.ParseInt(steamID, 10, 64)
	user, err := cc.db.GetUserByID(ID)
	if err != nil {
		return err
	}

	avatar := templates.Avatar(steamID, user.Avatar, user.Frame)

	return renderSVG(c, avatar)
}

func downloadFile(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
