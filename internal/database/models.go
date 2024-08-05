package database

import "database/sql"

type User struct {
	ID          int64          `json:"id"` // Steam64 ID
	DisplayName string         `json:"display_name"`
	VanityURL   sql.NullString `json:"vanity_url"`
	Avatar      sql.NullString `json:"avatar"`
	Frame       sql.NullString `json:"frame"`
	CreatedAt   string         `json:"created_at"`
	UpdateAt    sql.NullString `json:"update_at"`
}

type Query struct {
	Query     string `json:"query"`
	IP        string `json:"ip"`
	Country   string `json:"country"`
	CreatedAt string `json:"created_at"`
}
