package database

import (
	"github.com/paultag/go-dictd/dictd"
	"github.com/syndtr/goleveldb/leveldb"
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

	databaseBackend.storeDefinition("foo", "foo is a word that means something")

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
func (this *LevelDBDatabase) Match(name string, query string, strat string) []*dictd.Definition {
	return make([]*dictd.Definition, 0)
}

/*
 *
 */
func (this *LevelDBDatabase) Define(name string, query string) []*dictd.Definition {
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
func (this *LevelDBDatabase) storeDefinition(word string, def string) error {
	return this.db.Put([]byte(word), []byte(def), nil)
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
