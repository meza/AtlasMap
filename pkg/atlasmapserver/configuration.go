package atlasmapserver

import (
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/securecookie"
)

// Configuration options for the server.
type Configuration struct {
	Host               string
	Port               uint16
	StaticProxy        string
	StaticDir          string
	DisableCommands    bool
	FetchRateInSeconds int

	AtlasRedisAddress  string
	AtlasRedisPassword string
	AtlasRedisDB       int

	AdminSteamIDs []string

	// FS Store for now, may change to redis
	SessionStore string
	SessionKey   string
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func (s *AtlasMapServer) loadConfig() error {
	s.config = &Configuration{}

	port, err := strconv.ParseUint(getEnv("PORT", "3000"), 10, 16)
	if err != nil {
		return err
	}
	s.config.Port = uint16(port)

	s.config.DisableCommands, err = strconv.ParseBool(getEnv("DISABLECOMMANDS", "true"))
	if err != nil {
		return err
	}

	s.config.FetchRateInSeconds, err = strconv.Atoi(getEnv("FETCHRATE", "15"))
	if err != nil {
		return err
	}

	s.config.StaticDir = getEnv("STATICDIR", "")
	s.config.StaticProxy = getEnv("STATICPROXY", "")

	s.config.SessionStore = getEnv("SESSION_PATH", "./store")
	s.config.SessionKey = getEnv("SESSION_KEY", string(securecookie.GenerateRandomKey(32)))

	s.config.AtlasRedisAddress = getEnv("ATLAS_REDIS_ADDRESS", "localhost:6379")
	s.config.AtlasRedisPassword = getEnv("ATLAS_REDIS_PASSWORD", "")
	s.config.AtlasRedisDB, err = strconv.Atoi(getEnv("ATLAS_REDIS_DB", "0"))
	if err != nil {
		return err
	}

	s.config.AdminSteamIDs = strings.Split(getEnv("ADMIN_STEAMID_LIST", ""), " ")

	return nil
}
