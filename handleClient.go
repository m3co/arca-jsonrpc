package jsonrpc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
)

// RegisterSource stores the hierarchy of the handlers agains their
// context and method. The reason of registering the source is because
// we want to split this hierarchy into two parts. This part corresponds
// to the set of JSON-RPs from the client side.
func (s *Server) RegisterSource(
	method string, context string, rp RemoteProcedure) {
	if s.registersSource[context] == nil {
		rps := make(map[string]RemoteProcedure)
		rps[method] = rp
		s.registersSource[context] = rps
	} else {
		s.registersSource[context][method] = rp
	}
}

// RegisterTarget stores the hierarchy of the handlers agains their
// context and method. The reason of registering the source is because
// we want to split this hierarchy into two parts. This part corresponds
// to the set of JSON-RPs from the client side.
func (s *Server) RegisterTarget(
	method string, context string, rp DBRemoteProcedure) {
	if s.registersTarget[context] == nil {
		rps := make(map[string]DBRemoteProcedure)
		rps[method] = rp
		s.registersTarget[context] = rps
	} else {
		s.registersTarget[context][method] = rp
	}
}

// handleClient listens for any messages from conn and process it by using
// the method ProcessRequest
func (s *Server) handleClient(conn *net.Conn) {
	defer (*conn).Close()

	scanner := bufio.NewScanner(*conn)
	for scanner.Scan() {
		raw := scanner.Bytes()
		if len(raw) == 0 {
			continue
		}
		var request Request

		if err := json.Unmarshal(raw, &request); err != nil {
			log.Println("handleClient", err)
			base := &Base{}
			(*s).sendError(conn, base, &Error{
				Message: "Parse error",
				Code:    -32700,
				Data:    fmt.Sprint(err),
			})
			continue
		}

		s.ProcessRequest(&request, conn)
	}
	log.Println("disconnected")
}
