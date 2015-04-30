package database

import (
	"strings"

	"github.com/paultag/go-dictd/dictd"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"

	"github.com/jamesturk/go-jellyfish"
)

/*
 *
 */
func NewLevelDBDatabase(path string, description string) (dictd.Database, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}

	databaseBackend := LevelDBDatabase{
		description: description,
		db:          db,
	}

	return &databaseBackend, nil
}

/*
 *
 */
type LevelDBDatabase struct {
	dictd.Database

	description string
	db          *leveldb.DB
}

/*
 *
 */
func (this *LevelDBDatabase) Match(name string, query string, strat string) (defs []*dictd.Definition) {
	query = strings.ToLower(query)
	var results []string

	switch strat {
	case "prefix":
		results = this.scanPrefix(query)
	case "levenshtein":
		results = this.scanLevenshtein(query, 1)
	}

	for _, el := range results {
		def := &dictd.Definition{
			DictDatabase:     this,
			DictDatabaseName: name,
			Word:             el,
		}
		defs = append(defs, def)
	}

	return
}

/*
 *
 */
func (this *LevelDBDatabase) Define(name string, query string) []*dictd.Definition {
	query = strings.ToLower(query)
	data, err := this.db.Get([]byte(query), nil)
	if err != nil {
		/* If we don't have the key, let's bail out. */
		return make([]*dictd.Definition, 0)
	}
	els := make([]*dictd.Definition, 1)
	els[0] = &dictd.Definition{
		DictDatabase:     this,
		DictDatabaseName: name,
		Word:             query,
		Definition:       string(data),
	}
	return els
}

/*
 *
 */
func (this *LevelDBDatabase) Info(name string) string {
	return "Foo"
}

/*
 *
 */
func (this *LevelDBDatabase) Description(name string) string {
	return this.description
}

/*
 *
 *     MATCHERS
 *
 *
 *
 *
 */

/*
 */
func (this *LevelDBDatabase) scanLevenshtein(query string, threshold int) (ret []string) {
	iter := this.db.NewIterator(nil, nil)
	for iter.Next() {
		key := string(iter.Key())
		distance := jellyfish.Levenshtein(query, key)
		if distance <= threshold {
			/* XXX: Return ordered by distance? */
			ret = append(ret, key)
		}
	}
	iter.Release()
	return
}

/*
 */
func (this *LevelDBDatabase) scanPrefix(query string) (ret []string) {
	iter := this.db.NewIterator(util.BytesPrefix([]byte(query)), nil)

	for iter.Next() {
		word := string(iter.Key())
		ret = append(ret, word)
	}
	iter.Release()
	return
}
