package main

import (
	"log"

	"github.com/antihax/AtlasMap/atlasadminserver"
)

func main() {
	s := atlasadminserver.NewAtlasAdminServer()
	if err := s.Run(); err != nil {
		log.Fatalln(err)
	}
	log.Println("Server quit!")
}
