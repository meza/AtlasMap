package atlasdb

// GetAllEntities returns a hashmap of all entities. Limited to 4096 results
func (s *AtlasDB) GetAllEntities() (map[string]map[string]string, error) {
	return s.scanHash("entityinfo:*", 4096)
}
