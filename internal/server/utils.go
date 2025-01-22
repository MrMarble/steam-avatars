package server

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/mrmarble/steam-avatars/internal/steam"
)

func downloadFile(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func downloadFrame(c *steam.Client, steamID string) (string, error) {
	frame, err := c.GetAvatarFrame(steamID)
	if err != nil {
		return "", err
	}

	frameFile, err := downloadFile(frame)
	if err != nil {
		return "", fmt.Errorf("failed to download frame: %w", err)
	}

	return fmt.Sprintf("data:image/apng;base64,%s", base64.StdEncoding.EncodeToString(frameFile)), nil
}

func donwloadAvatar(c *steam.Client, player *steam.Player) (string, error) {
	avatar, err := c.GetAnimatedAvatar(player.SteamID)
	if err != nil {
		return "", fmt.Errorf("failed to get animated avatar for %q: %w", player.SteamID, err)
	}

	if avatar == "" {
		avatarFile, err := downloadFile(player.AvatarFull)
		if err != nil {
			return "", fmt.Errorf("failed to download avatar: %w", err)
		}
		avatar = fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(avatarFile))
	} else {
		avatarFile, err := downloadFile(avatar)
		if err != nil {
			return "", fmt.Errorf("failed to download animated avatar: %w", err)
		}
		avatar = fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(avatarFile))
	}

	return avatar, nil
}
