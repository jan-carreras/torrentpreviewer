package main

import (
	"log"
	"prevtorrent/cmd/http/bootstrap"
)

func main() {
	if err := bootstrap.Run(); err != nil {
		log.Fatal(err)
	}
}
