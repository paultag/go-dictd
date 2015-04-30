package main

import (
	"log"
	"net"

	"github.com/paultag/go-dictd/database"
	"github.com/paultag/go-dictd/dictd"
)

func main() {
	server := dictd.NewServer("pault.ag")
	jargon, err := database.NewLevelDBDatabase(
		"/home/tag/jargon.ldb",
		"jargon file",
	)
	congress, err := database.NewLevelDBDatabase(
		"/home/tag/congress.ldb",
		"congress words",
	)
	debian, err := database.NewLevelDBDatabase(
		"/home/tag/debian.ldb",
		"debian words",
	)
	if err != nil {
		log.Fatal(err)
	}

	urbanDB := database.UrbanDictionaryDatabase{}

	server.RegisterDatabase(jargon, "jargon")
	server.RegisterDatabase(congress, "congress")
	server.RegisterDatabase(debian, "debian")
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
