package main

import (
	"log"

	"github.com/antihax/AtlasMap/pkg/atlasmapserver"
)

func main() {
	s := atlasmapserver.NewAtlasMapServer()
	if err := s.Run(); err != nil {
		log.Fatalln(err)
	}
	log.Println("Server quit!")
}
