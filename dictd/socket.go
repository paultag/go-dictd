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

import (
	"strconv"
	"time"
)

/* socket.go - transport-aware interface for the dict protocol.
 *
 * In particular, this file contains a net/textproto based interface
 * to the rest of the go-dictd project, as well as parsing code for
 * the incoming requests. */

import (
	"errors"
	"log"
	"net"
	"net/textproto"
	"strings"
)

/* Session contains state data that lasts throughout the connection. */
type Session struct {
	MsgId      string
	Client     string
	Connection *textproto.Conn
	DictServer *Server
	Options    map[string]bool
}

/* Take an incoming line `line`, and split it according to the command
 * line spec in the RFC. */
func tokenizeLine(line string) ([]string, error) {
	var tokens []string
	chunks := strings.Split(line, " ")

	for _, v := range chunks {
		el := strings.Trim(v, " \t\n\r")
		if el != "" {
			tokens = append(tokens, el)
		}
	}

	if len(tokens) == 0 {
		return nil, errors.New("Busted command")
	}

	return tokens, nil
}

/* Parse an incoming line, and return a `dict.Command` suitable for
 * passing to internal (or external) handlers. */
func parseLine(line string) (*Command, error) {
	tokens, err := tokenizeLine(line)
	if err != nil {
		return nil, err
	}
	command := Command{
		Command: strings.ToUpper(tokens[0]),
		Params:  tokens[1:],
	}
	return &command, nil
}

/* Given a dict.Session and a dict.Command, route the command to the proper
 * handler, and dispatch the command. */
func handleCommand(session *Session, command *Command) {
	log.Printf("Incomming command from %s: %s", session.MsgId, command.Command)
	handler := session.DictServer.GetHandler(command)
	if handler == nil {
		unknownCommandHandler(session, *command)
	} else {
		handler(session, *command)
	}
}

/* Helper for commands to write out a text block */
func WriteTextBlock(session *Session, stream string) {
	if session.Options["MIME"] {
		session.Connection.Writer.PrintfLine(
			"Content-type: text/plain; charset=utf-8\n" +
				"Content-transfer-encoding: 8bit\n",
		)

	}

	writer := session.Connection.Writer.DotWriter()
	writer.Write([]byte(stream))
	writer.Close()
}

/* Helper for commands to write out a code line */
func WriteCode(session *Session, code int, message string) {
	session.Connection.Writer.PrintfLine("%d %s", code, message)
}

func generateMsgId(server *Server) string {
	return strconv.FormatInt(time.Now().UnixNano(), 10) +
		"." +
		"0000" +
		"@" +
		server.Name
}

/* Given a `dict.Server` and a `net.Conn`, do a bringup, and run the
 * `ReadLine` loop, dispatching commands to the correct internals. */
func Handle(server *Server, conn net.Conn) {
	proto := textproto.NewConn(conn)

	session := Session{
		MsgId:      generateMsgId(server),
		Client:     "",
		Connection: proto,
		DictServer: server,
		Options:    map[string]bool{},
	}

	session.Options["MIME"] = false /* Requiredish */

	/* Right, so we've got a connection, let's send the 220 and let the
	 * client know we're happy. */
	handshakeHandler(&session)

	for {
		line, err := proto.ReadLine()
		if err != nil {
			log.Printf("Error: %s", err)
			/* Usually an EOF */
			return
		}
		command, err := parseLine(line)
		if err != nil {
			log.Printf("Error: %s", err)
			continue
		}
		handleCommand(&session, command)
	}
}
