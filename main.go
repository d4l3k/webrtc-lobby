package main

import (
	"log"

	"github.com/d4l3k/webrtc-lobby/lobby"
)

func main() {
	s := lobby.NewServer()
	log.Fatal(s.Listen(":5000"))
}
