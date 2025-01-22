package database

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/valkey-io/valkey-go"
)

func (db *Database) GetUserByID(id int64) (*User, error) {
	var users []*User
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	if err := valkey.DecodeSliceOfJSON(db.client.Do(ctx, db.client.B().Mget().Key(strconv.Itoa(int(id))).Build()), &users); err != nil {
		return nil, err
	}

	if len(users) == 0 || users[0] == nil {
		return nil, valkey.Nil
	}

	return users[0], nil
}

func (db *Database) GetUserByVanityURL(vanity_url string) (*User, error) {
	ctx := context.Background()
	userID, err := db.client.Do(ctx, db.client.B().Get().Key(vanity_url).Build()).ToString()
	if err != nil {
		return nil, err
	}

	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, err
	}

	return db.GetUserByID(id)
}

func (db *Database) CreateUser(user *User) error {
	ctx := context.Background()
	// Store vanity URL to ID mapping
	if err := db.client.Do(ctx, db.client.B().Set().Key(user.VanityURL).Value(strconv.Itoa(int(user.ID))).Ex(time.Hour*24).Build()).Error(); err != nil {
		return fmt.Errorf("failed to store vanity URL to ID mapping: %w", err)
	}
	if err := db.client.Do(ctx, db.client.B().Set().Key(strconv.Itoa(int(user.ID))).Value(valkey.JSON(user)).Ex(time.Hour*24).Build()).Error(); err != nil {
		return fmt.Errorf("failed to store user: %w", err)
	}

	return nil
}
