package database

type User struct {
	ID          int64  `json:"id"` // Steam64 ID
	DisplayName string `json:"display_name"`
	VanityURL   string `json:"vanity_url"`
	Avatar      string `json:"avatar"`
	Frame       string `json:"frame"`
}
