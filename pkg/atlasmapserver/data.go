package atlasmapserver

import "errors"

func (s *AtlasMapServer) GetPlayerIDFromSteamID(steamID string) (int64, error) {
	playerID, ok := s.mapSteamIDPlayerID.Load(steamID)
	if ok {
		return playerID.(int64), nil
	}
	return 0, errors.New("cannot locate playerID")
}

func (s *AtlasMapServer) GetSteamIDFromPlayerID(playerID int64) (string, error) {
	steamID, ok := s.mapSteamIDPlayerID.Load(playerID)
	if ok {
		return steamID.(string), nil
	}
	return "", errors.New("cannot locate steamID")
}
