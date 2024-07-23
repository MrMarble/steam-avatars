package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type GetAvatarFrameResponse struct {
	Response struct {
		AvatarFrame struct {
			AppID           int    `json:"appid"`
			CommunityItemID string `json:"communityitemid"`
			ImageLarge      string `json:"image_large"`
			ImageSmall      string `json:"image_small"`
			Name            string `json:"name"`
		} `json:"avatar_frame"`
	} `json:"response"`
}

type GetAnimatedAvatarResponse struct {
	Response struct {
		AvatarFrame struct {
			AppID           int    `json:"appid"`
			CommunityItemID string `json:"communityitemid"`
			ImageLarge      string `json:"image_large"`
			ImageSmall      string `json:"image_small"`
			Name            string `json:"name"`
		} `json:"avatar"`
	} `json:"response"`
}

type ResolveVanityURLResponse struct {
	Response struct {
		SteamID string `json:"steamid"`
		Success int    `json:"success"`
	} `json:"response"`
}

type GetPlayerSummariesResponse struct {
	Response struct {
		Players []struct {
			SteamID     string `json:"steamid"`
			AvatarFull  string `json:"avatarfull"`
			RealName    string `json:"realname"`
			PersonaName string `json:"personaname"`
			ProfileURL  string `json:"profileurl"`
		} `json:"players"`
	} `json:"response"`
}

type GetMiniProfileBackgroundResponse struct {
	Response struct {
		ProfileBackground struct {
			AppID           int    `json:"appid"`
			CommunityItemID string `json:"communityitemid"`
			ImageLarge      string `json:"image_large"`
			MovieWebm       string `json:"movie_webm"`
			MovieMP4        string `json:"movie_mp4"`
		} `json:"profile_background"`
	} `json:"response"`
}

func GetSteamID(client *http.Client, steamAPIKey, vanityURL string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.steampowered.com/ISteamUser/ResolveVanityURL/v1/?key=%s&vanityurl=%s", steamAPIKey, vanityURL), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Failed to get SteamID: %s", err)
	}
	defer resp.Body.Close()

	var data ResolveVanityURLResponse

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", fmt.Errorf("Failed to get SteamID: %s", err)
	}

	return data.Response.SteamID, nil
}

func GetAvatarFrame(client *http.Client, steamAPIKey, steamID string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.steampowered.com/IPlayerService/GetAvatarFrame/v1/?key=%s&steamid=%s", steamAPIKey, steamID), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data GetAvatarFrameResponse

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("https://cdn.akamai.steamstatic.com/steamcommunity/public/images/%s", data.Response.AvatarFrame.ImageSmall), nil
}

func GetAnimatedAvatar(client *http.Client, steamAPIKey, steamID string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.steampowered.com/IPlayerService/GetAnimatedAvatar/v1/?key=%s&steamid=%s", steamAPIKey, steamID), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data GetAnimatedAvatarResponse

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", err
	}

	if data.Response.AvatarFrame.ImageSmall == "" {
		return "", nil
	}

	return fmt.Sprintf("https://cdn.akamai.steamstatic.com/steamcommunity/public/images/%s", data.Response.AvatarFrame.ImageSmall), nil
}

func GetAvatar(client *http.Client, steamAPIKey, steamID string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/?key=%s&steamids=%s", steamAPIKey, steamID), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data GetPlayerSummariesResponse

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", err
	}

	return data.Response.Players[0].AvatarFull, nil
}

func GetMiniProfileBackground(client *http.Client, steamAPIKey, steamID string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.steampowered.com/IPlayerService/GetMiniProfileBackground/v1/?key=%s&steamid=%s", steamAPIKey, steamID), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data GetMiniProfileBackgroundResponse

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("https://cdn.akamai.steamstatic.com/steamcommunity/public/images/%s", data.Response.ProfileBackground.MovieMP4), nil
}

func isSteamID(name string) bool {
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
