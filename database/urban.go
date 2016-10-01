package database

import (
	"log"

	"github.com/jessfraz/udict/api"
	"pault.ag/go/dictd/dictd"
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
	response, err := api.Define(query)
	if err != nil {
		log.Printf("Error getting from UD: %s", err)
		return
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
	return "Look up words on http://www.urbandictionary.com/"
}

/*
 *
 */
func (this *UrbanDictionaryDatabase) Description(name string) string {
	return "Urban Dictonary"
}

/*
 *
 */
func (this *UrbanDictionaryDatabase) Strategies(name string) map[string]string {
	return map[string]string{}
}
