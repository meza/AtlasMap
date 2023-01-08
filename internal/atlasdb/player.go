package atlasdb

import (
	"context"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

type PlayerServerInfo struct {
	PlayerID          string `redis:"PlayerId"`
	CurrentServerID   int64  `redis:"CurrentServerId"`
	FirstHomeServerID int64  `redis:"FirstHomeServerId"`
	LogLineID         int64  `redis:"LogLineId"`
}

// GetPlayerSteamID returns the SteamID of a playerID.
func (s *AtlasDB) GetPlayerServerInfoFromSteamID(ctx context.Context, steamID string) (*PlayerServerInfo, error) {
	p := &PlayerServerInfo{}
	if err := s.db.HGetAll(ctx, "playerserverinfo:"+steamID).Scan(p); err != nil {
		return nil, err
	}
	return p, nil
}

type PlayerInfo struct {
	PlayerID     int64  `redis:"PlayerId"`
	TribeID      int64  `redis:"TribeID"`
	PlayerName   string `redis:"PlayerName"`
	LastOnlineAt int64  `redis:"LastOnlineAt"`
	RankGroupID  int    `redis:"RankGroupId"`
}

// GetPlayerSteamID returns the SteamID of a playerID.
func (s *AtlasDB) GetPlayerInfoFromPlayerID(ctx context.Context, playerID int64) (*PlayerInfo, error) {
	p := &PlayerInfo{}
	if err := s.db.HGetAll(ctx, "playerinfo:"+strconv.FormatInt(playerID, 10)).Scan(p); err != nil {
		return nil, err
	}
	return p, nil
}

// GetPlayerSteamID returns the SteamID of a playerID.
func (s *AtlasDB) GetAllPlayerID(ctx context.Context) ([]int64, error) {
	p := []int64{}
	r := s.db.Keys(ctx, "PlayerDataId:*")
	players, err := r.Result()
	if err != nil {
		return nil, err
	}

	for _, player := range players {
		s := strings.Split(player, ":")
		if len(s) == 2 {
			id, err := strconv.ParseInt(s[1], 10, 64)
			if err != nil {
				log.Error().Err(err).Msgf("processing", player)
				continue
			}
			p = append(p, id)
		}
	}
	return p, nil
}

// GetSteamIDFromPlayerID returns the SteamID of a playerID.
func (s *AtlasDB) GetSteamIDFromPlayerID(ctx context.Context, playerID int64) (string, error) {
	var p string
	if err := s.db.Get(ctx, "PlayerDataId:"+strconv.FormatInt(playerID, 10)).Scan(&p); err != nil {
		return p, err
	}
	return p, nil
}

// GetPlayerSteamID returns the SteamID of a playerID.
func (s *AtlasDB) GetPlayerBySteamID(ctx context.Context, playerID string) (string, error) {
	return s.db.Get(ctx, "PlayerDataId:"+playerID).Result()
}
