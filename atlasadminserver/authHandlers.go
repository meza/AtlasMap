package atlasadminserver

import (
	"net/http"

	"github.com/prometheus/common/log"
	"github.com/solovev/steam_go"
)

func (s *AtlasAdminServer) loginHandler(w http.ResponseWriter, r *http.Request) {
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
