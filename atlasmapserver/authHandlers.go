package atlasmapserver

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/prometheus/common/log"
	"github.com/solovev/steam_go"
)

func (s *AtlasMapServer) logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := s.store.Get(r, "atlas-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err)
		return
	}

	// Force deletion of the cookie and session
	session.Options.MaxAge = -1

	// Save changes
	if err := session.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err)
		return
	}

	http.Redirect(w, r, "/", 301)
}

func (s *AtlasMapServer) loginHandler(w http.ResponseWriter, r *http.Request) {
	opID := steam_go.NewOpenId(r)
	switch opID.Mode() {
	case "":
		http.Redirect(w, r, opID.AuthUrl(), 301)
	case "cancel":
		http.Redirect(w, r, "/", 301)
	default:
		steamID, err := opID.ValidateAndGetId()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error(err)
			return
		}
		session, err := s.store.Get(r, "atlas-session")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error(err)
			return
		}
		session.Values["steamID"] = steamID

		if err := session.Save(r, w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error(err)
			return
		}
		http.Redirect(w, r, "/", 301)
	}
}

// Determine if the request is an administrator
func (s *AtlasMapServer) isAdmin(r *http.Request) bool {
	session, err := s.store.Get(r, "atlas-session")
	if err != nil {
		return false
	}
	steamID, ok := session.Values["steamID"].(string)
	if ok {
		for _, id := range s.config.AdminSteamIDs {
			if steamID == id {
				return true
			}
		}
	}
	return false
}

// Determine if the request is a tribe administrator
func (s *AtlasMapServer) isTribeAdmin(r *http.Request) bool {
	session, err := s.store.Get(r, "atlas-session")
	if err != nil {
		return false
	}
	steamID, ok := session.Values["steamID"].(string)
	if ok {
		tribe, err := s.getTribeDataFromSteamID(steamID)
		if err != nil {
			return false
		}
		admins, ok := tribe["TribeAdmins"]
		if ok {
			for _, admin := range strings.Split(strings.Trim(admins, "()[]"), " ") {
				if steamID == admin {
					return true
				}
			}
		}
	}

	return false
}

func (s *AtlasMapServer) accountHandler(w http.ResponseWriter, r *http.Request) {
	session, err := s.store.Get(r, "atlas-session")

	steamID, ok := session.Values["steamID"].(string)
	if ok {
		player, err := s.getPlayerDataFromSteamID(steamID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error(err)
			return
		}

		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(player)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error(err)
			return
		}
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err)
		return
	}
}
