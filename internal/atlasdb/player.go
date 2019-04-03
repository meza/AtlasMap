package atlasdb

// GetAllPlayers returns a hashmap of all entities. Limited to 80000 results
func (s *AtlasDB) GetAllPlayers() (map[string]map[string]string, error) {
	return s.scanHash("playerinfo:*", 80000)
}

// GetPlayerSteamID returns the SteamID of a playerID.
func (s *AtlasDB) GetPlayerSteamID(playerID string) (string, error) {
	return s.db.Get("PlayerDataId:" + playerID).Result()
}

// GetReverseSteamIDMap returns a map of all steamID mapped to playerID. Limited to 80000 results
func (s *AtlasDB) GetReverseSteamIDMap() (map[string]string, error) {
	players, err := s.scanString("PlayerDataId:*", 80000)
	if err != nil {
		return nil, err
	}
	return reverseMap(players), nil
}
