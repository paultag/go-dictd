package main

import (
	"log"
	"os"
	"strings"

	"pault.ag/go/dictd/database"
	"pault.ag/go/dictd/format"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Give me a path to the db and a path to a file")
	}

	path := os.Args[1]
	dbFile := os.Args[2]

	defs := format.ParseJargonFormat(dbFile)
	db, err := database.NewLevelDBDatabase(path, "")

	if err != nil {
		log.Fatal(err)
	}

	for _, def := range defs {
		word := strings.ToLower(def.Word)
		db.WriteDefinition(word, def.Definition)
	}

}
