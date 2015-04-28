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

/* commands.go - protocol handlers for the core functions of the dict protocol
 *
 * This contains the core protocol implementaion, including syntax error
 * handlers, the initial handshake, and the core MUST-haves */

import (
	"strings"
)

/*
 *
 */
func unknownCommandHandler(session *Session, command Command) {
	WriteCode(session, 500, "unknown command")
}

/*
 *
 */
func handshakeHandler(session *Session) {
	session.Connection.Writer.PrintfLine("220 %s <%s> <%s>",
		"go-dictd",
		"mime",
		session.MsgId,
	)
}

/*
 */
func clientCommandHandler(session *Session, command Command) {
	/* Ignore everything for now, in the future, we should be
	 * setting the value of the message over into session.Client or
	 * something. */
	WriteCode(session, 250, "ok")
}

/*
 */
func showCommandHandler(session *Session, command Command) {
	/* SHOW DB
	 * SHOW DATABASES
	 * SHOW STRAT
	 * SHOW STRATEGIES
	 * SHOW INFO database
	 * SHOW SERVER */

	if len(command.Params) < 1 {
		syntaxErrorHandler(session, command)
		return
	}

	param := strings.ToUpper(command.Params[0])

	switch param {
	case "DB", "DATABASES":
		session.Connection.Writer.PrintfLine(
			"110 %d database(s) present",
			len(session.DictServer.databases),
		)
		for db := range session.DictServer.databases {
			databaseBackend := session.DictServer.GetDatabase(db)
			session.Connection.Writer.PrintfLine(
				"%s \"%s\"",
				db,
				databaseBackend.Description(),
			)
		}
		session.Connection.Writer.PrintfLine(".")
		WriteCode(session, 250, "ok")
		return
	case "STRAT", "STRATEGIES":
		session.Connection.Writer.PrintfLine("111 2 present")
		WriteTextBlock(session, `exact "Match"
prefix "Prefix"`)
		session.Connection.Writer.PrintfLine("250 ok")
		return
	case "INFO":
		if len(command.Params) < 2 {
			syntaxErrorHandler(session, command)
			return
		}
		name := command.Params[1]
		session.Connection.Writer.PrintfLine("112 information for %s", name)
		databaseBackend := session.DictServer.GetDatabase(name)

		if databaseBackend == nil {
			WriteCode(session, 550, "invalid database")
			return
		}
		WriteTextBlock(session, databaseBackend.Info())
		WriteCode(session, 250, "ok")
		return
	case "SERVER":
		WriteCode(session, 114, "server information")
		WriteTextBlock(session, session.DictServer.Info)
		WriteCode(session, 250, "ok")
		return
	}

	unknownCommandHandler(session, command)

}

/*
 *
 */
func optionCommandHandler(session *Session, command Command) {

	if len(command.Params) < 1 {
		syntaxErrorHandler(session, command)
		return
	}

	param := strings.ToUpper(command.Params[0])

	switch param {
	case "MIME":
		session.Options["MIME"] = !session.Options["MIME"]
		if session.Options["MIME"] {
			WriteCode(session, 250, "ok - mime I guess")
		} else {
			WriteCode(session, 250, "ok - no mime I guess")
		}
		return
	}

	unknownCommandHandler(session, command)
}

/*
 */
func syntaxErrorHandler(session *Session, command Command) {
	WriteCode(session, 501, "syntax error, illegal parameters")
}

/*
 */
func quitCommandHandler(session *Session, command Command) {
	WriteCode(session, 221, "bye")
	session.Connection.Close()
}

/*
 */
func writeDefinition(
	session *Session,
	databaseBackend Database,
	definition *Definition,
	database string,
) {
	session.Connection.Writer.PrintfLine(
		"151 \"%s\" %s \"%s\"",
		definition.Word,
		database,
		databaseBackend.Description(),
	)
	WriteTextBlock(session, definition.Definition)
}

/*
 */
func defineCommandHandler(session *Session, command Command) {

	if len(command.Params) <= 1 {
		syntaxErrorHandler(session, command)
		return
	}

	database := command.Params[0]
	word := command.Params[1]

	/*
	 * Dispatch on ! or * for those behaviors
	 */

	databaseBackend := session.DictServer.GetDatabase(database)
	if databaseBackend == nil {
		WriteCode(session, 550, "invalid database")
		return
	}

	words := databaseBackend.Define(word)
	session.Connection.Writer.PrintfLine(
		"150 %d definitions retrieved",
		len(words),
	)

	for _, el := range words {
		writeDefinition(session, databaseBackend, el, database)
	}
	WriteCode(session, 250, "ok")
}

/*
 *
 */
func registerDefaultHandlers(server *Server) {
	server.RegisterHandler("CLIENT", clientCommandHandler)
	server.RegisterHandler("DEFINE", defineCommandHandler)
	server.RegisterHandler("OPTION", optionCommandHandler)
	server.RegisterHandler("SHOW", showCommandHandler)
	server.RegisterHandler("QUIT", quitCommandHandler)
}
