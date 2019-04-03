package atlasdb

// GetAllTribes returns a hashmap of all tribes. Limited to 4096 results
func (s *AtlasDB) GetAllTribes() (map[string]map[string]string, error) {
	return s.scanHash("tribedata:*", 4096)
}
