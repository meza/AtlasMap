// Package atlasdb handles the redis data flow with the Atlas Redis server
package atlasdb

import "github.com/go-redis/redis"

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
	_, err := s.db.Ping().Result()
	if err != nil {
		return nil, err
	}

	return s, nil
}
