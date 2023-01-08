package atlasmapserver

import (
	"github.com/gorilla/mux"
)

func (s *AtlasMapServer) apiRouter(r *mux.Route) {
	//	router := r.Subrouter()

}

/*
// sendCommand publishes an event to the GeneralNotifications:GlobalCommands
// redis PubSub channel. To send a server command, prepend "ID::X,Y::" where
// ID is the packed server ID; X and Y are the relative lng and lat locations.
func (s *AtlasMapServer) sendCommand(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != "POST" || s.config.DisableCommands {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		w.WriteHeader(http.StatusOK)
		return
	}

	encoded := fmt.Sprintf("%s", body)

	log.Println("publish:", encoded)
	result, err := s.redisClient.Publish("GeneralNotifications:GlobalCommands", encoded).Result()
	if err != nil {
		log.Println("redis error for: ", encoded, "; ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write([]byte(strconv.FormatInt(result, 10)))
}
*/
