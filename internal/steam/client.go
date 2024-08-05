package steam

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	baseURL  = "https://api.steampowered.com"
	assetURL = "https://cdn.akamai.steamstatic.com/steamcommunity/public/images/"
)

type Client struct {
	apiKey string
	c      *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		c: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) get(url string, params map[string]string, v interface{}) error {
	// params
	url += fmt.Sprintf("?key=%s", c.apiKey)
	if len(params) > 0 {
		for key, value := range params {
			url += fmt.Sprintf("&%s=%s", key, value)
		}
	}
	req, err := http.NewRequest("GET", baseURL+url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.c.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(v)
}

func (c *Client) GetSteamID(vanityURL string) (string, error) {
	if IsSteamID(vanityURL) {
		return vanityURL, nil
	}

	var data ResolveVanityURLResponse
	err := c.get("/ISteamUser/ResolveVanityURL/v1/", map[string]string{"vanityurl": vanityURL}, &data)
	if err != nil {
		return "", err
	}
	if data.Response.Success != 1 {
		return "", fmt.Errorf("vanity url not found")
	}

	return data.Response.SteamID, nil
}

func (c *Client) GetAvatarFrame(steamID string) (string, error) {
	var data GetAvatarFrameResponse
	err := c.get("/IPlayerService/GetAvatarFrame/v1/", map[string]string{"steamid": steamID}, &data)
	if err != nil {
		return "", err
	}

	if data.Response.AvatarFrame.ImageSmall == "" {
		return "", nil
	}

	return fmt.Sprintf("%s%s", assetURL, data.Response.AvatarFrame.ImageSmall), nil

}

func (c *Client) GetAnimatedAvatar(steamID string) (string, error) {
	var data GetAnimatedAvatarResponse
	err := c.get("/IPlayerService/GetAnimatedAvatar/v1/", map[string]string{"steamid": steamID}, &data)
	if err != nil {
		return "", err
	}

	if data.Response.AvatarFrame.ImageSmall == "" {
		return "", nil
	}

	return fmt.Sprintf("%s%s", assetURL, data.Response.AvatarFrame.ImageSmall), nil
}

func (c *Client) GetPlayer(steamID string) (*Player, error) {
	var data GetPlayerSummariesResponse
	err := c.get("/ISteamUser/GetPlayerSummaries/v2/", map[string]string{"steamids": steamID}, &data)
	if err != nil {
		return nil, err
	}

	return &data.Response.Players[0], nil
}

func IsSteamID(name string) bool {
	if len(name) != 17 {
		return false
	}

	for _, c := range name {
		if c < '0' || c > '9' {
			return false
		}
	}

	return true
}
