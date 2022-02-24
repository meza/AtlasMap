// Package atlasdb handles the redis data flow with the Atlas Redis server
package atlasdb

import (
	"context"

	"github.com/go-redis/redis/v8"
)

// AtlasDB provides an interface to the Atlas DB
type AtlasDB struct {
	db *redis.Client
}

// NewAtlasDB provides a new DB pool
func NewAtlasDB(address string, password string, db int) (*AtlasDB, error) {
	s := &AtlasDB{
		db: redis.NewClient(&redis.Options{
			Addr:     address,
			Password: password,
			DB:       db,
		}),
	}

	// Test connection
	_, err := s.db.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return s, nil
}
