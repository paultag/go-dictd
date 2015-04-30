package main

import (
	"encoding/json"
	"log"
	"net"
	"os"

	"github.com/paultag/go-dictd/database"
	"github.com/paultag/go-dictd/dictd"
)

/* Used for the JSON file that we can load off of */
type Configuration struct {
	Name string
	Info string

	Databases []struct {
		Name string
		Path string
		Desc string
	}
}

/* Given a config, load them up! */
func loadDatabases(path string) (config *Configuration, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}

	decoder := json.NewDecoder(file)
	config = new(Configuration)
	err = decoder.Decode(&config)

	if err != nil {
		return nil, err
	}

	return
}

func main() {
	config, err := loadDatabases("/etc/dictd.json")
	if err != nil {
		log.Fatal(err)
	}

	server := dictd.NewServer(config.Name)

	for _, dbConfig := range config.Databases {
		db, err := database.NewLevelDBDatabase(
			dbConfig.Path,
			dbConfig.Desc,
		)
		if err != nil {
			log.Fatal(err)
		}
		server.RegisterDatabase(db, dbConfig.Name)
	}

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
