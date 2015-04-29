package main

import (
	"log"
	"net"

	"github.com/paultag/go-dictd/database"
	"github.com/paultag/go-dictd/dictd"
)

func main() {
	server := dictd.NewServer("pault.ag")
	levelDB, err := database.NewLevelDBDatabase("/home/tag/jargon.ldb", "jargon file")
	if err != nil {
		log.Fatal(err)
	}

	urbanDB := database.UrbanDictionaryDatabase{}

	server.RegisterDatabase(levelDB, "jargon")
	server.RegisterDatabase(&urbanDB, "urban")

	link, err := net.Listen("tcp", ":2628")
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := link.Accept()
		if err != nil {
			log.Printf("Error: %s", err)
		}
		go dictd.Handle(&server, conn)
	}
}
