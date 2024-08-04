package steam

type GetAvatarFrameResponse struct {
	Response struct {
		AvatarFrame struct {
			AppID           int    `json:"appid"`
			CommunityItemID string `json:"communityitemid"`
			ImageLarge      string `json:"image_large"`
			ImageSmall      string `json:"image_small"` // This is the URL to the animated frame
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
		Players []Player `json:"players"`
	} `json:"response"`
}

type Player struct {
	SteamID     string `json:"steamid"`
	AvatarFull  string `json:"avatarfull"`
	RealName    string `json:"realname"`
	PersonaName string `json:"personaname"`
	ProfileURL  string `json:"profileurl"`
}
