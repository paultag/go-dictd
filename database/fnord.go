package database

import (
	"github.com/paultag/go-dictd/dictd"
)

type FnordDatabase struct {
	dictd.Database
}

func (this *FnordDatabase) Match(query string, strat string) []*dictd.Definition {
	els := make([]*dictd.Definition, 1)
	els[0] = &dictd.Definition{Word: query}
	return els
}

func (this *FnordDatabase) Define(query string) []*dictd.Definition {
	els := make([]*dictd.Definition, 1)
	els[0] = &dictd.Definition{
		Word:       query,
		Definition: query,
	}
	return els
}

func (this *FnordDatabase) Info() string {
	return `Some long
Description goes here

and here.`
}

func (this *FnordDatabase) Description() string {
	return "fake testing database"
}
