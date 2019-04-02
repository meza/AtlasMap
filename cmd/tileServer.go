package main

import (
	"log"

	"github.com/antihax/AtlasMap/atlasadminserver"
)

func main() {
	s, err := atlasadminserver.NewAtlasAdminServer()
	if err != nil {
		log.Fatalln(err)
	}
	if err = s.Run(); err != nil {
		log.Fatalln(err)
	}
	log.Println("Server quit!")
}
