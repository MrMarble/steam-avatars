package server

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mrmarble/steam-avatars/internal/database"
	"github.com/mrmarble/steam-avatars/internal/server/templates"
	"github.com/mrmarble/steam-avatars/internal/steam"
)

func handleIndex(c echo.Context) error {
	cc := c.(*Context)
	users, err := cc.db.GetLatestUsers()
	if err != nil {
		c.Logger().Error(err)
	}
	page := templates.Index(users)
	return renderView(c, page)
}

func handleSearch(c echo.Context) error {
	cc := c.(*Context)
	name := c.FormValue("name")
	if name == "" {
		return c.JSON(400, map[string]string{"error": "name is required"})
	}

	c.Logger().Info("searching for vanity URL ", name)

	user, err := cc.db.GetUserByVanityOrID(name)
	if err != nil && err != sql.ErrNoRows {
		return err
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
	return renderView(c, templates.Result(strID, user.Avatar.String, user.Frame.String, c.Request().URL.Scheme+"://"+c.Request().Host+"/avatar/"+strID))
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
	frameFile, err := downloadFile(frame)
	if err != nil {
		return nil, err
	}
	frame = fmt.Sprintf("data:image/apng;base64,%s", base64.StdEncoding.EncodeToString(frameFile))

	avatar, err := c.GetAnimatedAvatar(steamID)
	if err != nil {
		return nil, err
	}

	if avatar == "" {
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
	}

	ID, _ := strconv.ParseInt(steamID, 10, 64)

	return &database.User{
		ID:          ID,
		VanityURL:   sql.NullString{String: query, Valid: true},
		DisplayName: summary.PersonaName,
		Avatar:      sql.NullString{String: avatar, Valid: true},
		Frame:       sql.NullString{String: frame, Valid: true},
		CreatedAt:   time.Now().Format(time.RFC3339),
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

	avatar := templates.Avatar(steamID, user.Avatar.String, user.Frame.String)

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
