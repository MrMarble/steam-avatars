package database

func (db *Database) GetUserByID(id int) (*User, error) {
	user := &User{}
	err := db.db.QueryRow("SELECT * FROM users WHERE id = ?", id).Scan(&user.ID, &user.Username, &user.DisplayName, &user.VanityURL, &user.Avatar, &user.Frame)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (db *Database) GetUserByVanityURL(vanity_url string) (*User, error) {
	user := &User{}
	err := db.db.QueryRow("SELECT * FROM users WHERE vanity_url = ?", vanity_url).Scan(&user.ID, &user.Username, &user.DisplayName, &user.VanityURL, &user.Avatar, &user.Frame)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (db *Database) CreateUser(user *User) error {
	_, err := db.db.Exec("INSERT INTO users (id, username, display_name, vanity_url, avatar, frame) VALUES (?, ?, ?, ?, ?, ?)", user.ID, user.Username, user.DisplayName, user.VanityURL, user.Avatar, user.Frame)
	if err != nil {
		return err
	}

	return nil
}
