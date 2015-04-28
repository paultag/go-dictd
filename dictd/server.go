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

package dictd

/* server.go - infrastructure code for the dictd server
 *
 * In particular, this file contains a few interfaces and commonly used
 * structs for message passing, as well as generic routing code. */

import (
	"errors"
)

/* Command is the encapsulation for a user's request of the Server. */
type Command struct {
	Command string
	Params  []string
}

/* Definition is the encapsulation of a response for a given Entry. */
type Definition struct {
	Word             string
	Definition       string
	DictDatabase     Database
	DictDatabaseName string
}

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
}

/* Server encapsulation.
 *
 * This contains a bundle of useful helpers, as well as a few data structures
 * to handle registered Databases and Commands. */
type Server struct {
	Name      string
	Info      string
	databases map[string]Database
	commands  map[string]func(*Session, Command)
}

/* Define a word against the server, according to fun rules! */
func (this *Server) Match(
	database string,
	query string,
	strat string,
) ([]*Definition, error) {
	/* Right, so we've been asked to figure out what a word is.
	 * The RFC has special handling based on the database name,
	 * so we're going to go ahead and figure out what we should
	 * be doing here. */

	switch database {
	case "!":
		/* The RFC states that we search all Databases for entries, and
		 * if we hit, we should return all Definitions for the given word
		 * for that Database. */
		for database, databaseBackend := range this.databases {
			defs := databaseBackend.Match(database, query, strat)
			if defs != nil && len(defs) != 0 {
				return defs, nil
			}
		}
		return make([]*Definition, 0), nil

	case "*":
		/* The RFC states that we search all Databases for entries, and
		 * return *all* Definitions for the given word for all Databases. */
		var allDefs = make([]*Definition, 0)
		for database, databaseBackend := range this.databases {
			defs := databaseBackend.Match(database, query, strat)
			allDefs = append(allDefs, defs...)
		}
		return allDefs, nil
	}

	/* Otherwise, let's go with the boring usual behavior -- try to get
	 * the database, and return defs for that one DB. */
	db := this.GetDatabase(database)
	if db == nil {
		return nil, errors.New("No such database")
	}
	return db.Match(database, query, strat), nil
}

/* Define a word against the server, according to fun rules! */
func (this *Server) Define(
	database string,
	query string,
) ([]*Definition, error) {

	/* Right, so we've been asked to figure out what a word is.
	 * The RFC has special handling based on the database name,
	 * so we're going to go ahead and figure out what we should
	 * be doing here. */

	switch database {
	case "!":
		/* The RFC states that we search all Databases for entries, and
		 * if we hit, we should return all Definitions for the given word
		 * for that Database. */
		for database, databaseBackend := range this.databases {
			defs := databaseBackend.Define(database, query)
			if defs != nil && len(defs) != 0 {
				return defs, nil
			}
		}
		return make([]*Definition, 0), nil

	case "*":
		/* The RFC states that we search all Databases for entries, and
		 * return *all* Definitions for the given word for all Databases. */
		var allDefs = make([]*Definition, 0)
		for database, databaseBackend := range this.databases {
			defs := databaseBackend.Define(database, query)
			allDefs = append(allDefs, defs...)
		}
		return allDefs, nil
	}

	/* Otherwise, let's go with the boring usual behavior -- try to get
	 * the database, and return defs for that one DB. */
	db := this.GetDatabase(database)
	if db == nil {
		return nil, errors.New("No such database")
	}
	return db.Define(database, query), nil
}

/* Register dict.Database `database` under `name`. */
func (this *Server) RegisterDatabase(database Database, name string) {
	this.databases[name] = database
}

/* Get dict.Database that has been registered under `name`. */
func (this *Server) GetDatabase(name string) Database {
	if value, ok := this.databases[name]; ok {
		return value
	}
	return nil
}

/* Register a Command `handler` under name `name`. */
func (this *Server) RegisterHandler(
	name string,
	handler func(*Session, Command),
) {
	this.commands[name] = handler
}

/* Get a Command handler for the given dict.Command `command` */
func (this *Server) GetHandler(command *Command) func(*Session, Command) {
	name := command.Command

	if value, ok := this.commands[name]; ok {
		return value
	}
	return nil
}

/* Create a new server by name `name`. */
func NewServer(name string) Server {
	server := Server{
		Name:      name,
		Info:      "",
		commands:  map[string]func(*Session, Command){},
		databases: map[string]Database{},
	}
	registerDefaultHandlers(&server)
	return server
}
