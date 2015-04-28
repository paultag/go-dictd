package main

import (
	"log"
	"net"

	"github.com/paultag/go-dictd/database"
	"github.com/paultag/go-dictd/dictd"
)

func main() {
	server := dictd.NewServer("pault.ag")
	server.RegisterDatabase(&database.FnordDatabase{}, "test")

	link, err := net.Listen("tcp", ":2017")
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
