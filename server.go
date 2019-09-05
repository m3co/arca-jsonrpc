package jsonrpc

import (
	"net"
	"sync"
)

// Close takes the listen and close channel and closes them
func (s *Server) Close() {
	(*s.listen).Close()
}

// Broadcast sends to all the active connections the given message
func (s *Server) Broadcast(msg []byte) {
	for _, conn := range s.conns {
		s.write(conn, msg)
	}
}

// BroadcastError takes a JSON-RPC error and sends it to all connections
func (s *Server) BroadcastError(base *Base, response *Error) {
	for _, conn := range s.conns {
		s.sendError(conn, base, response)
	}
}

// Start prepares and launches the json-rpc server
func (s *Server) Start(ready *chan bool) (err error) {
	listen, err := net.Listen("tcp", s.Address)
	if err != nil {
		*ready <- false
		return err
	}

	s.plugBlocker = &sync.Mutex{}
	s.writeBlocker = &sync.Mutex{}
	s.conns = make([]*net.Conn, 0)
	s.listen = &listen
	s.registersSource = make(map[string]map[string]RemoteProcedure)
	s.registersTarget = make(map[string]map[string]DBRemoteProcedure)

	go (func() {
		for {
			conn, err := listen.Accept()
			if err != nil {
				return
			}
			s.plug(&conn)
			go (func(conn *net.Conn) {
				s.handleClient(conn)
				s.unplug(conn)
			})(&conn)
		}
	})()

	*ready <- true
	return nil
}
