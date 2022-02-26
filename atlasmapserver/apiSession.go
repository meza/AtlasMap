package atlasmapserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/antihax/AtlasMap/internal/atlasdb"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/rs/zerolog/log"
)

func (s *AtlasMapServer) sessionRouter(r *mux.Route) {
	router := r.Subrouter()
	router.Use(s.sessionMiddleware)
	router.HandleFunc("/account", s.accountHandler)
	router.HandleFunc("/events", s.eventHandler)
}

// sessionMiddleware adds session data to the context
func (s *AtlasMapServer) sessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := s.store.Get(r, "session")
		if err != nil {
			log.Error().Err(err).Msg("bad session")
			http.SetCookie(w, &http.Cookie{Name: "session", MaxAge: -1, Path: "/"})
			return
		}

		_, ok := session.Values["steamID"].(string)
		if !ok {
			http.Error(w, "Not authenticated", http.StatusUnauthorized)
			return
		}
		_, ok = session.Values["playerID"].(int64)
		if !ok {
			http.Error(w, "Not authenticated", http.StatusUnauthorized)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), SessionKey, session))
		next.ServeHTTP(w, r)
	})
}

type accountData struct {
	Tribe        *atlasdb.TribeData
	Player       *atlasdb.PlayerInfo
	PlayerServer *atlasdb.PlayerServerInfo
}

func (s *AtlasMapServer) accountHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	var err error

	session := r.Context().Value(SessionKey).(*sessions.Session)
	steamID := session.Values["steamID"].(string)

	accData := accountData{}

	playerID := session.Values["playerID"].(int64)
	accData.PlayerServer, err = s.db.GetPlayerServerInfoFromSteamID(r.Context(), steamID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error().Err(err).Msg("db.GetPlayerServerInfoFromSteamID")
		return
	}

	accData.Player, err = s.db.GetPlayerInfoFromPlayerID(r.Context(), playerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error().Err(err).Msg("db.GetPlayerInfoFromPlayerID")
		return
	}

	if accData.Player.TribeID > 0 {
		accData.Tribe, err = s.db.GetTribeByID(r.Context(), accData.Player.TribeID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error().Err(err).Msg("db.GetTribeByID")
			return
		}
	}

	// output json struct
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(accData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error().Err(err).Msg("accountHandler json encode")
		return
	}
}

func (s *AtlasMapServer) eventHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	session := r.Context().Value(SessionKey).(*sessions.Session)
	steamID, ok := session.Values["steamID"].(string)
	if !ok {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}
	playerID, ok := session.Values["playerID"].(int64)
	if !ok {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	playerInfo, err := s.db.GetPlayerInfoFromPlayerID(r.Context(), playerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error().Err(err).Msg("db.GetPlayerInfoFromPlayerID")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	channel := s.broker.AddUser(steamID, playerInfo.TribeID)

	defer s.broker.RemoveChannel(channel)
	for {
		select {
		case msg := <-channel:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-r.Context().Done():
			log.Debug().Msgf("eventHandler %s", r.Context().Err())
			flusher.Flush()
			return
		}
	}
}
