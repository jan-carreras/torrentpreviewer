package main

import (
	"log"
	"prevtorrent/cmd/prevtorrent/bootstrap"
)

func main() {
	err := bootstrap.Run()
	if err != nil {
		log.Fatal(err)
	}
}
