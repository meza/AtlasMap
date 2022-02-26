package atlasmapserver

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/antihax/AtlasMap/atlasmapserver/eventbroker"
	"github.com/antihax/AtlasMap/internal/atlasdb"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"net/http"
	"sync"
	"time"
)

// AtlasMapServer provides administrative services to an Atlas Cluster over http
type AtlasMapServer struct {
	// Map steamID to playerID as redis does not hold this in a useful manner
	mapSteamIDPlayerID sync.Map
	mapPlayerIDSteamID sync.Map

	broker *eventbroker.EventBroker

	config *Configuration
	router *mux.Router
	db     *atlasdb.AtlasDB

	// Session store and CSRF protection
	store *sessions.FilesystemStore
}

// NewAtlasMapServer creates a new server
func NewAtlasMapServer() *AtlasMapServer {
	return &AtlasMapServer{
		router: mux.NewRouter(),
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
func (s *AtlasMapServer) Run() error {
	// Load configuration from environment
	if err := s.loadConfig(); err != nil {
		return err
	}

	// Setup session store
	s.store = sessions.NewFilesystemStore(s.config.SessionStore, []byte(s.config.SessionKey))
	s.store.MaxAge(2400)
	// Setup our DB pool
	db, err := atlasdb.NewAtlasDB(
		s.config.AtlasRedisAddress,
		s.config.AtlasRedisPassword,
		s.config.AtlasRedisDB,
	)
	if err != nil {
		return err
	}
	s.db = db

	s.broker = eventbroker.NewEventBroker(db)

	// Poll the database for data
	go s.fetch()

	// API Endpoints
	s.apiRouter(s.router.PathPrefix("/api/"))
	s.sessionRouter(s.router.PathPrefix("/s/"))

	// Login endpoints
	s.router.HandleFunc("/login", s.loginHandler)
	s.router.HandleFunc("/logout", s.logoutHandler)

	// Serve static content
	s.router.PathPrefix("/").Handler(http.FileServer(http.Dir(s.config.StaticDir)))

	endpoint := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	log.Info().Msgf("listening on %s", endpoint)
	return http.ListenAndServe(endpoint, s.router)
}

func (s *AtlasMapServer) fetch() {

	throttle := time.NewTicker(time.Duration(s.config.FetchRateInSeconds) * time.Second)

	for {
		// Get all players and update maps to include new players
		playerIDList, err := s.db.GetAllPlayerID(context.Background())
		if err != nil {
			log.Error().Err(err).Msg("db.GetAllPlayerID")
			continue
		}
		for _, playerID := range playerIDList {
			if _, ok := s.mapPlayerIDSteamID.Load(playerID); !ok {
				// fetch from redis
				steamID, err := s.db.GetSteamIDFromPlayerID(context.Background(), playerID)
				if err != nil {
					log.Error().Err(err).Msg("db.GetSteamIDFromPlayerID")
					continue
				}
				s.mapPlayerIDSteamID.Store(playerID, steamID)
				s.mapSteamIDPlayerID.Store(steamID, playerID)
			}
		}

		<-throttle.C
	}
}
