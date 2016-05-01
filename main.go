package main

import (
	"log"

	"github.com/d4l3k/webrtc-lobby/lobby"
)

func main() {
	s, err := lobby.NewServer()
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(s.Listen(":5000"))
}
