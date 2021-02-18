package main

import (
	"log"
	"prevtorrent/cmd/cli/torrentprev/bootstrap"
)

func main() {
	err := bootstrap.Run()
	if err != nil {
		log.Fatal(err)
	}
}
