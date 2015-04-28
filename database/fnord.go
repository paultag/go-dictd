package database

import (
	"github.com/paultag/go-dictd/dictd"
)

type FnordDatabase struct{ dictd.Database }

func (this *FnordDatabase) Match(name string, query string, strat string) []*dictd.Definition {
	els := make([]*dictd.Definition, 1)
	els[0] = &dictd.Definition{
		DictDatabase:     this,
		DictDatabaseName: name,
		Word:             query,
	}
	return els
}

func (this *FnordDatabase) Define(name string, query string) []*dictd.Definition {
	els := make([]*dictd.Definition, 1)
	els[0] = &dictd.Definition{
		DictDatabase:     this,
		DictDatabaseName: name,
		Word:             query,
		Definition:       query,
	}
	return els
}

func (this *FnordDatabase) Info(name string) string {
	return `Some long
Description goes here

and here.`
}

func (this *FnordDatabase) Description(name string) string {
	return "fake testing database"
}
