package database

func (db *Database) GetUserByID(id int64) (*User, error) {
	user := &User{}
	err := db.db.QueryRow("SELECT * FROM users WHERE id = ?", id).Scan(&user.ID, &user.DisplayName, &user.VanityURL, &user.Avatar, &user.Frame, &user.CreatedAt, &user.UpdateAt)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (db *Database) GetUserByVanityURL(vanity_url string) (*User, error) {
	user := &User{}
	err := db.db.QueryRow("SELECT * FROM users WHERE vanity_url = ?", vanity_url).Scan(&user.ID, &user.DisplayName, &user.VanityURL, &user.Avatar, &user.Frame, &user.CreatedAt, &user.UpdateAt)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (db *Database) GetUserByVanityOrID(query string) (*User, error) {
	user := &User{}
	err := db.db.QueryRow("SELECT * FROM users WHERE vanity_url = ? OR id = ?", query, query).Scan(&user.ID, &user.DisplayName, &user.VanityURL, &user.Avatar, &user.Frame, &user.CreatedAt, &user.UpdateAt)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (db *Database) CreateUser(user *User) error {
	_, err := db.db.Exec("INSERT INTO users (id, display_name, vanity_url, avatar, frame, created_at) VALUES (?, ?, ?, ?, ?, ?)", user.ID, user.DisplayName, user.VanityURL, user.Avatar, user.Frame, user.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) GetLatestUsers() ([]User, error) {
	rows, err := db.db.Query("SELECT * FROM users ORDER BY updated_at,created_at DESC LIMIT 10")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := make([]User, 0)
	for rows.Next() {
		user := User{}
		err := rows.Scan(&user.ID, &user.DisplayName, &user.VanityURL, &user.Avatar, &user.Frame, &user.CreatedAt, &user.UpdateAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}
