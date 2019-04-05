package atlasadminserver

import "errors"

func (s *AtlasAdminServer) getPlayerDataFromSteamID(steamID string) (map[string]string, error) {
	s.steamDataLock.RLock()
	playerID, ok := s.steamData[steamID]
	s.steamDataLock.RUnlock()
	if !ok {
		return nil, errors.New("No pathfinder found for SteamID")
	}
	s.playerDataLock.RLock()
	defer s.playerDataLock.RUnlock()
	return s.playerData[playerID], nil
}
