package atlasadminserver

import (
	"strconv"
)

// Configuration options for the server.
type Configuration struct {
	Host               string
	Port               uint16
	TerritoryURL       string
	StaticDir          string
	DisableCommands    bool
	FetchRateInSeconds int
	RedisAddress       string
	RedisPassword      string
	RedisDB            int
}

func (s *AtlasAdminServer) loadConfig() error {
	s.config = &Configuration{}

	port, err := strconv.ParseUint(getEnv("PORT", "8880"), 10, 16)
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

	s.config.TerritoryURL = getEnv("TERRITORY_URL", "http://localhost:8881/territoryTiles/")
	s.config.StaticDir = getEnv("STATICDIR", "./www")

	s.config.RedisAddress = getEnv("REDIS_ADDRESS", "localhost:6379")
	s.config.RedisPassword = getEnv("REDIS_PASSWORD", "foobared")
	s.config.RedisDB, err = strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		return err
	}

	return nil
}
