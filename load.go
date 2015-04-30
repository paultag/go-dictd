package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/paultag/go-dictd/format"
	"github.com/syndtr/goleveldb/leveldb"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Give me a path to the db and a path to a file")
	}

	path := os.Args[1]
	dbFile := os.Args[2]

	defs := format.ParseJargonFormat(dbFile)

	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	for _, def := range defs {
		word := strings.ToLower(def.Word)
		fmt.Printf("%s\n", word)
		db.Put([]byte(word), []byte(def.Definition), nil)
	}

}
