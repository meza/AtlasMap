package atlasdb

import (
	"encoding/binary"
	"strconv"
	"strings"

	"github.com/go-redis/redis"
)

// ServerID unpacks the packed server ID. Each Server has an X and Y ID which
// corresponds to its 2D location in the game world. The ID is packed into
// 32-bits as follows:
//   +--------------+--------------+
//   | X (uint16_t) | Y (uint16_t) |
//   +--------------+--------------+
func ServerID(packed string) (split [2]uint16, err error) {
	var id uint64
	id, err = strconv.ParseUint(packed, 10, 32)
	if err != nil {
		return
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(id))
	split[0] = binary.LittleEndian.Uint16(buf[:2])
	split[1] = binary.LittleEndian.Uint16(buf[2:])
	return
}

// Call HGETALL on a pattern and return a map
func (s *AtlasDB) scanHash(pattern string, maxItems int64) (map[string]map[string]string, error) {
	records := make(map[string]map[string]string)

	// [GSG] Scan is slower than Keys but provides gaps for other things to execute
	var keys []string
	iter := s.db.Scan(0, pattern, maxItems).Iterator()
	for iter.Next() {
		keys = append(keys, iter.Val())
	}

	// If the iterator has an error, do not continue to pull the hash
	if err := iter.Err(); err != nil {
		return nil, err
	}

	// Build a pipeline of requests for each key
	results := make(map[string]*redis.StringStringMapCmd)
	pipe := s.db.Pipeline()
	for _, id := range keys {
		results[id] = pipe.HGetAll(id)
	}

	// Execute all requests
	if _, err := pipe.Exec(); err != nil {
		return nil, err
	}

	// Build a map of the results
	for _, id := range keys {
		key := id
		if strings.Contains(id, ":") {
			parts := strings.Split(id, ":")
			key = parts[1]
		}
		var err error
		records[key], err = results[id].Result()
		if err != nil {
			return nil, err
		}
	}

	return records, nil
}

func reverseMap(m map[string]string) map[string]string {
	n := make(map[string]string)
	for k, v := range m {
		n[v] = k
	}
	return n
}

// Call GET on a pattern and return a map
func (s *AtlasDB) scanString(pattern string, maxItems int64) (map[string]string, error) {
	records := make(map[string]string)

	// [GSG] Scan is slower than Keys but provides gaps for other things to execute
	var keys []string
	iter := s.db.Scan(0, pattern, maxItems).Iterator()
	for iter.Next() {
		keys = append(keys, iter.Val())
	}

	// If the iterator has an error, do not continue to pull the hash
	if err := iter.Err(); err != nil {
		return nil, err
	}

	// Build a pipeline of requests for each key
	results := make(map[string]*redis.StringCmd)
	pipe := s.db.Pipeline()
	for _, id := range keys {
		results[id] = pipe.Get(id)
	}

	// Execute all requests
	if _, err := pipe.Exec(); err != nil {
		return nil, err
	}

	// Build a map of the results
	for _, id := range keys {
		var err error
		key := id
		if strings.Contains(id, ":") {
			parts := strings.Split(id, ":")
			key = parts[1]
		}
		records[key], err = results[id].Result()
		if err != nil {
			return nil, err
		}
	}

	return records, nil
}
