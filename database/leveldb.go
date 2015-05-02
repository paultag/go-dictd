/**
 * Copyright (c) Paul R. Tagliamonte, 2015
 *
 * Permission is hereby granted, free of charge, to any person obtaining a
 * copy of this software and associated documentation files (the "Software"),
 * to deal in the Software without restriction, including without limitation
 * the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the
 * Software is furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 * THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
 * FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
 * DEALINGS IN THE SOFTWARE. */

package database

/* leveldb.go - leveldb backend for serving dictd dictionaries.
 *
 * This backend is a super fast and hecka neato backend that uses Google's
 * leveldb to store and optimize seeking values to dispatch back to the user.
 *
 * leveldb is a simple key value store, so we've had to build some magic
 * on top of it. Basically, I needed to build in key namespacing, so I was
 * able to make some assumptions based on the dictd RFC2229.
 *
 * Keys are structured as:
 *
 *  "{namespace}\n{key}", since \n is never valid in a search query (even if
 *  it was, we split onece on the first, soo, whatever :) )
 *
 * The actual definitions are in "\nword", all lower case.
 *
 * We also build up indexes using this, by storing the precomputed / rendered
 * lookup strings in a namespace. Something like soundex would be
 * "soundex\n{key}". We can then do the magic over the incoming word and do
 * an O(1) lookup on that key. Magic, mirite. */

import (
	"sort"
	"strings"

	"github.com/paultag/go-dictd/dictd"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"

	"github.com/jamesturk/go-jellyfish"
)

/* Create a new LevelDB Database. `path` should be the full filesystem path
 * to the leveldb database, and `description` should be what we tell the user
 * the database is when they ask about it. Something short, for `SHOW DB`. */
func NewLevelDBDatabase(path string, description string) (*LevelDBDatabase, error) {
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

/* LevelDB Database container. This contains some fun bits (like the
 * leveldb.DB object, and the description the user gave us over in
 * NewLevelDBDatabase. */
type LevelDBDatabase struct {
	dictd.Database

	description string
	db          *leveldb.DB
}

/* Handle incoming RFC2229 MATCH requests.
 *
 * Currently supported MATCH algorithms:
 *
 *  [default] - Metaphone
 *            - Prefix    (byte prefixes)
 *            - Soundex
 *            - Levenshtein
 */
func (this *LevelDBDatabase) Match(name string, query string, strat string) (defs []*dictd.Definition) {
	query = strings.ToLower(query)
	var results []string

	switch strat {
	case "metaphone", ".":
		results = this.matchMetaphone(query)
	case "prefix":
		results = this.scanPrefix(query)
	case "soundex":
		results = this.matchSoundex(query)
	case "anagram":
		results = this.matchAnagram(query)
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

/* Handle incoming `DEFINE` calls. */
func (this *LevelDBDatabase) Define(name string, query string) []*dictd.Definition {
	query = strings.ToLower(query)
	data, err := this.get("", query)
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

/* Get all valid Strategies */
func (this *LevelDBDatabase) Strategies(name string) map[string]string {
	return map[string]string{
		"levenshtein": "Levenshtein distance",
		"soundex":     "Soundex matches",
		"metaphone":   "Metaphone matches",
		"anagram":     "Anagram matches",
	}
}

/* Handle the information call (SHOW INFO `name`) for this database. */
func (this *LevelDBDatabase) Info(name string) string {
	return "Foo"
}

/* Handle the short description of what this database does (for
 * inline `SHOW DB` output) */
func (this *LevelDBDatabase) Description(name string) string {
	return this.description
}

/* DB Specific calls below */

/*
 * Write a "namespaced" key into the LevelDB Database.
 *
 * `namespace` must never contain a newline.
 */
func (this *LevelDBDatabase) write(namespace string, key string, value string) {
	query := namespace + "\n" + key
	this.db.Put([]byte(query), []byte(value), nil)
}

/*
 * Get a "namespaced" key out of the LevelDB Database.
 *
 * `namespace` must never contain a newline.
 */
func (this *LevelDBDatabase) get(namespace string, key string) (value string, err error) {
	data, err := this.db.Get([]byte(namespace+"\n"+key), nil)
	return string(data), err
}

/* Given the namespace `namespace`, and key `key`, write the `word` in the
 * internal "list" syntax for the index. */
func (this *LevelDBDatabase) writeIndex(namespace string, key string, word string) {
	var values []string

	data, err := this.get(namespace, key)

	if err != nil {
		values = []string{}
	} else {
		/* Values are newline delimed */
		values = strings.Split(string(data), "\n")
	}

	for _, el := range values {
		if el == word {
			return
		}
	}

	values = append(values, word)
	this.write(namespace, key, strings.Join(values, "\n"))
}

/*
 *
 */
func sortString(word string) string {
	sorted := strings.Split(word, "")
	sort.Strings(sorted)
	return strings.Join(sorted, "")
}

/*
 * Given a word `word`, defined by definition `definition`, write this out
 * to the LevelDB database, and generate all Indexes we need.
 *
 */
func (this *LevelDBDatabase) WriteDefinition(word string, definition string) {
	/* Right, now let's build up indexes on the word */

	this.write("", word, definition) /* no namespace for words */

	/* Hilarious. */
	this.writeIndex("anagram", sortString(word), word)

	/* Right, now let's build up some indexes */
	this.writeIndex("soundex", jellyfish.Soundex(word), word)

	if len(word) > 2 { /* Fixme? */
		metaWords := jellyfish.Metaphone(word)

		/* FO BA BAR BAZ */
		for _, el := range strings.Split(metaWords, " ") {
			this.writeIndex("metaphone", el, word)
		}
	}

}

/*  MATCHERS  */

/* Scan the index for Levenshtein matches. Since we need the target string
 * and the query string, this requires an O(n) scan of the index. This kinda
 * sucks. This might actually be the least performant MATCH algorithm. */
func (this *LevelDBDatabase) scanLevenshtein(query string, threshold int) (ret []string) {
	iter := this.db.NewIterator(util.BytesPrefix([]byte("\n")), nil)
	for iter.Next() {
		key := string(iter.Key())[1:]
		distance := jellyfish.Levenshtein(query, key)
		if distance <= threshold {
			/* XXX: Return ordered by distance? */
			ret = append(ret, key)
		}
	}
	iter.Release()
	return
}

/* Scan the index for matches based on the first few bytes. Since we need
 * to scan for the incoming query, we iterate over the chunk of the DB
 * we need. LevelDB is sorted alphabetically, we can actually just get the
 * part of the DB prefixed by using the leveldb.util.BytesPrefix helper. */
func (this *LevelDBDatabase) scanPrefix(query string) (ret []string) {
	query = "\n" + query /* See namespacing code */

	iter := this.db.NewIterator(util.BytesPrefix([]byte(query)), nil)

	for iter.Next() {
		word := string(iter.Key())[1:]
		ret = append(ret, word)
	}
	iter.Release()
	return
}

/* Given a precomputed index and a key, do a lookup of the "array" we're storing
 * in LevelDB. Return a list of strings that are hits for that target. */
func (this *LevelDBDatabase) matchFromIndex(namespace string, key string) (ret []string) {
	data, err := this.get(namespace, key)
	if err != nil {
		return []string{}
	}
	return strings.Split(string(data), "\n")
}

/* Internal anagram matcher. */
func (this *LevelDBDatabase) matchAnagram(query string) (ret []string) {
	return this.matchFromIndex("anagram", sortString(query))
}

/* Internal soundex matcher. */
func (this *LevelDBDatabase) matchSoundex(query string) (ret []string) {
	return this.matchFromIndex("soundex", jellyfish.Soundex(query))
}

/* Internal metaphone matcher. */
func (this *LevelDBDatabase) matchMetaphone(query string) (ret []string) {
	meta := jellyfish.Metaphone(query)
	for _, el := range strings.Split(meta, " ") {
		ret = append(ret, this.matchFromIndex("metaphone", el)...)
	}

	/* right, so ret may have multiples */
	ordering := map[string]int{}
	for _, el := range ret {
		ordering[el] = 0 /* update this to count / sort */
	}

	r := []string{}
	for k, _ := range ordering {
		r = append(ret, k)
	}
	return r
}
