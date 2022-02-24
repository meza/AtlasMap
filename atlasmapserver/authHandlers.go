package atlasmapserver

import (
	"net/http"
	"strconv"

	"github.com/gorilla/sessions"
	"github.com/rs/zerolog/log"
	"github.com/solovev/steam_go"
)

type contextKey int

const (
	SessionKey contextKey = iota
)

func (s *AtlasMapServer) logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, ok := r.Context().Value(SessionKey).(*sessions.Session)
	if ok {
		// Force deletion of the cookie and session
		session.Options.MaxAge = -1

		// Save changes
		if err := session.Save(r, w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error().Err(err).Msg("save session")
			return
		}
	} else {
		// Delete cookie if we can't read the session
		s.clearSessionCookie(w)
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func (s *AtlasMapServer) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: "session", MaxAge: -1, Path: "/"})
}

func (s *AtlasMapServer) loginHandler(w http.ResponseWriter, r *http.Request) {
	opID := steam_go.NewOpenId(r)
	switch opID.Mode() {
	case "":
		http.Redirect(w, r, opID.AuthUrl(), http.StatusMovedPermanently)
	case "cancel":
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	default:
		steamID, err := opID.ValidateAndGetId()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error().Err(err).Msg("validate steam OpenID")
			return
		}

		// sanity steamID
		_, err = strconv.ParseUint(steamID, 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error().Err(err).Msg("ParseInt")
			return
		}

		// Create a new session and store steamID and privileges
		session, err := s.store.New(r, "session")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error().Err(err).Msg("login new session")
			s.clearSessionCookie(w)
			return
		}

		// PlayerID should not change frequently
		playerID, err := s.GetPlayerIDFromSteamID(steamID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error().Err(err).Msg("GetPlayerIDFromSteamID")
			return
		}

		session.Values["steamID"] = steamID
		session.Values["playerID"] = playerID

		// Set administrator
		for _, id := range s.config.AdminSteamIDs {
			if steamID == id {
				session.Values["admin"] = true
				break
			}
		}

		// Save session and redirect to home
		if err := session.Save(r, w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error().Err(err).Msg("save session")
			return
		}
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	}
}

// Determine if the request is a tribe administrator
/*
func (s *AtlasMapServer) isTribeAdmin(r *http.Request) bool {
	session, err := s.store.Get(r, "session")
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
}*/
