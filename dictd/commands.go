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
	session.Connection.Writer.PrintfLine("500 %s", "unknown command")
}

/*
 *
 */
func handshakeHandler(session *Session) {
	session.Connection.Writer.PrintfLine("220 %s <%s> <%s>",
		"pault.ag dictd proto",
		"mime",
		session.MsgId,
	)
}

/*
 */
func clientCommandHandler(session *Session, command Command) {
	session.Connection.Writer.PrintfLine("250 %s", "ok")
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
			session.Connection.Writer.PrintfLine("250 ok - mime I guess")
		} else {
			session.Connection.Writer.PrintfLine("250 ok - no mime I guess")
		}
		return
	}

	session.Connection.Writer.PrintfLine("500 %s", "unknown command")
}

/*
 */
func syntaxErrorHandler(session *Session, command Command) {
	session.Connection.Writer.PrintfLine(
		"501 syntax error, illegal parameters",
	)
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
		session.Connection.Writer.PrintfLine("550 invalid database")
		return
	}

	words := databaseBackend.Define(word)
	session.Connection.Writer.PrintfLine(
		"150 %d definitions retrieved",
		len(words),
	)

	for _, el := range words {
		session.Connection.Writer.PrintfLine(
			"151 \"%s\" %s \"%s\"",
			word,
			database,
			databaseBackend.Description(),
		)
		writer := session.Connection.Writer.DotWriter()
		writer.Write([]byte(el.Definition))
		writer.Close()
	}
	session.Connection.Writer.PrintfLine("250 ok")
}

/*
 *
 */
func registerDefaultHandlers(server *Server) {
	server.RegisterHandler("CLIENT", clientCommandHandler)
	server.RegisterHandler("DEFINE", defineCommandHandler)
	server.RegisterHandler("OPTION", optionCommandHandler)
}
