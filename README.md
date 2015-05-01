go-dictd
========

`go-dictd` is a Go library for implementing custom `RFC2229`/`dictd` protocol
servers.

Basically, the coolest part is that you can define your own custom "tables"
for the dictd server. So far, this comes with a
[LevelDB](https://github.com/paultag/go-dictd/blob/master/database/leveldb.go)
and an
[Urban Dictionary](https://github.com/paultag/go-dictd/blob/master/database/urban.go)
example.


Writing a custom Database
-------------------------

The protocol you have to implement is the `dictd.Database` interface,
which looks something like:

```do
/* Database is an interface for external Database "Backends" to implement. */
type Database interface {

	/* Method to handle incoming `MATCH` commands. */
	Match(name string, query string, strat string) []*Definition

	/* Method to handle incoming `DEFINE` commands. */
	Define(name string, query string) []*Definition

	/* Method to handle incoming `SHOW INFO` commands. */
	Info(name string) string

	/* Method to return a one-line Description of the Database. */
	Description(name string) string

	/* Get a list of valid Match Strategies. */
	Strategies(name string) map[string]string
}
```
