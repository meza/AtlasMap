package atlasadminserver

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"

	"github.com/coreos/go-oidc"

	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis"
)

// AtlasAdminServer provides administrative services to an Atlas Cluster over http
type AtlasAdminServer struct {
	redisClient   *redis.Client
	gameData      map[string]EntityInfo
	gameDataLock  sync.RWMutex
	tribeData     map[string]string
	tribeDataLock sync.RWMutex
	config        *Configuration
	steamOpenID   *oidc.Provider
}

// NewAtlasAdminServer creates a new server
func NewAtlasAdminServer() *AtlasAdminServer {
	return &AtlasAdminServer{
		gameData:  make(map[string]EntityInfo),
		tribeData: make(map[string]string),
	}
}

// EntityInfo record in redis.
type EntityInfo struct {
	EntityID                string
	ParentEntityID          string
	EntityType              string
	EntitySubType           string
	EntityName              string
	TribeID                 string
	ServerXRelativeLocation float64
	ServerYRelativeLocation float64
	ServerID                [2]uint16
	LastUpdatedDBAt         uint64
	NextAllowedUseTime      uint64
}

// ServerLocation relative percentage to a specific server's origin.
type ServerLocation struct {
	ServerID                [2]uint16
	ServerXRelativeLocation float64
	ServerYRelativeLocation float64
}

// Run starts the server processing
func (s *AtlasAdminServer) Run() error {

	// Load configuration from environment
	if err := s.loadConfig(); err != nil {
		return err
	}

	// Setup redis connection
	s.redisClient = redis.NewClient(&redis.Options{
		Addr:     s.config.RedisAddress,
		Password: s.config.RedisPassword,
		DB:       s.config.RedisDB,
	})

	// Test connection
	_, err := s.redisClient.Ping().Result()
	if err != nil {
		return err
	}

	// Get an OpenID Provider
	s.steamOpenID, err = oidc.NewProvider(context.Background(), "https://steamcommunity.com/openid")
	if err != nil {
		return err
	}

	// Poll the database for data
	go s.fetch()

	// Setup handlers
	http.HandleFunc("/gettribes", s.getTribes)
	http.HandleFunc("/getdata", s.getData)
	http.HandleFunc("/travels", s.getPathTravelled)
	http.HandleFunc("/territoryURL", s.getTerritoryURL)
	http.Handle("/", http.FileServer(http.Dir(s.config.StaticDir)))

	// Don't serve command handler if disabled
	if !s.config.DisableCommands {
		http.HandleFunc("/command", s.sendCommand)
	}

	endpoint := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	log.Println("Listening on ", endpoint)
	return http.ListenAndServe(endpoint, nil)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func (s *AtlasAdminServer) fetch() {

	var kidsWithBadParents map[string]bool
	kidsWithBadParents = make(map[string]bool)

	throttle := time.NewTimer(time.Duration(s.config.FetchRateInSeconds) * time.Second)

	for {
		tribes := make(map[string]string)
		entities := make(map[string]EntityInfo)

		for _, record := range scan(s.redisClient, "tribedata:*") {
			tribes[record["TribeID"]] = record["TribeName"]
			fmt.Printf("%+v\n", record)
		}
		s.tribeDataLock.Lock()
		s.tribeData = tribes
		s.tribeDataLock.Unlock()

		for _, record := range scan(s.redisClient, "entityinfo:*") {
			// log.Println(id)
			info := newEntityInfo(record)
			entities[info.EntityID] = *info
			fmt.Printf("%+v\n", info)
		}

		// sanity check entity data, e.g. any missing parent ids?
		for k, v := range entities {
			if v.ParentEntityID != "0" {
				if _, parentFound := entities[v.ParentEntityID]; !parentFound {
					if _, dontSpamLog := kidsWithBadParents[k]; !dontSpamLog {
						kidsWithBadParents[k] = true
						log.Printf("Entity %s references parent %s that does not exist, removing from list", k, v.ParentEntityID)
					}
					delete(entities, k)
				}
			}
		}

		s.gameDataLock.Lock()
		s.gameData = entities
		s.gameDataLock.Unlock()

		<-throttle.C
	}
}

func scan(client *redis.Client, pattern string) map[string]map[string]string {
	records := make(map[string]map[string]string)

	start := time.Now()

	// Scan is slower than Keys but provides gaps for other things to execute
	var keys []string
	iter := client.Scan(0, pattern, 5000).Iterator()
	for iter.Next() {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		log.Fatal(err)
	}

	// Load each entity
	results := make(map[string]*redis.StringStringMapCmd)
	pipe := client.Pipeline()
	for _, id := range keys {
		results[id] = pipe.HGetAll(id)
	}
	pipe.Exec()
	for _, id := range keys {
		records[id], _ = results[id].Result()
	}

	elapsed := time.Since(start)
	log.Printf("Redis scan took %s", elapsed)

	return records
}

// serverID unpacks the packed server ID. Each Server has an X and Y ID which
// corresponds to its 2D location in the game world. The ID is packed into
// 32-bits as follows:
//   +--------------+--------------+
//   | X (uint16_t) | Y (uint16_t) |
//   +--------------+--------------+
func serverID(packed string) (split [2]uint16, err error) {
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

// newEntityInfo transforms a Key-Value record into a new EntityInfo object.
func newEntityInfo(record map[string]string) *EntityInfo {
	var info EntityInfo
	info.EntityID = record["EntityID"]
	info.ParentEntityID = record["ParentEntityID"]
	info.EntityName = record["EntityName"]
	info.EntityType = record["EntityType"]
	info.ServerXRelativeLocation, _ = strconv.ParseFloat(record["ServerXRelativeLocation"], 64)
	info.ServerYRelativeLocation, _ = strconv.ParseFloat(record["ServerYRelativeLocation"], 64)
	info.LastUpdatedDBAt, _ = strconv.ParseUint(record["LastUpdatedDBAt"], 10, 64)
	info.NextAllowedUseTime, _ = strconv.ParseUint(record["NextAllowedUseTime"], 10, 64)

	var ok bool
	var tmpTribeID string
	if tmpTribeID, ok = record["TribeID"]; !ok {
		tmpTribeID, _ = record["TribeId"]
	}
	info.TribeID = tmpTribeID

	var tmpServerID string
	if tmpServerID, ok = record["ServerID"]; !ok {
		tmpServerID, _ = record["ServerId"]
	}
	info.ServerID, _ = serverID(tmpServerID)

	// convert entity class to a subtype
	var tmpEntityClass string
	if tmpEntityClass, ok = record["EntityClass"]; !ok {
		tmpEntityClass = "none"
	}
	tmpEntityClass = strings.ToLower(tmpEntityClass)
	if strings.Contains(tmpEntityClass, "none") {
		info.EntitySubType = "None"
	} else if strings.Contains(tmpEntityClass, "brigantine") {
		info.EntitySubType = "Brigantine"
	} else if strings.Contains(tmpEntityClass, "dinghy") {
		info.EntitySubType = "Dingy"
	} else if strings.Contains(tmpEntityClass, "raft") {
		info.EntitySubType = "Raft"
	} else if strings.Contains(tmpEntityClass, "sloop") {
		info.EntitySubType = "Sloop"
	} else if strings.Contains(tmpEntityClass, "schooner") {
		info.EntitySubType = "Schooner"
	} else if strings.Contains(tmpEntityClass, "galleon") {
		info.EntitySubType = "Galleon"
	} else {
		info.EntitySubType = "None"
	}

	return &info
}
