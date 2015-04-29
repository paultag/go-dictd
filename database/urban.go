package database

import (
	"log"

	"github.com/jfrazelle/udict/api"
	"github.com/paultag/go-dictd/dictd"
)

/*
 *
 */
type UrbanDictionaryDatabase struct {
	dictd.Database
}

/*
 *
 */
func (this *UrbanDictionaryDatabase) Match(name string, query string, strat string) []*dictd.Definition {
	return []*dictd.Definition{}
}

/*
 *
 */
func (this *UrbanDictionaryDatabase) Define(name string, query string) (definitions []*dictd.Definition) {
	response, err := api.DefineWord(query)
	if err != nil {
		log.Printf("Error getting from UD: %s", err)
	}

	for _, el := range response.Results {
		definitions = append(definitions, &dictd.Definition{
			Word:             el.Word,
			Definition:       el.Definition,
			DictDatabase:     this,
			DictDatabaseName: name,
		})
		log.Printf("%s\n", el.Word)
	}
	return
}

/*
 *
 */
func (this *UrbanDictionaryDatabase) Info(name string) string {
	return "Foo"
}

/*
 *
 */
func (this *UrbanDictionaryDatabase) Description(name string) string {
	return "UD"
}
