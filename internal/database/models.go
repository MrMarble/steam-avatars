package database

type User struct {
	ID          int    `json:"id"` // Steam64 ID
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	VanityURL   string `json:"vanity_url"`
	Avatar      string `json:"avatar"`
	Frame       string `json:"frame"`
}
