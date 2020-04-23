package jsonrpc

import (
	"net"
	"sync"
)

// Close takes the listen and close channel and closes them
func (s *Server) Close() error {
	return s.listen.Close()
}

// Broadcast sends to all the active connections the given message
func (s *Server) Broadcast(msg []byte) {
	for _, conn := range s.conns {
		if err := s.write(conn, msg); err != nil {
			//log.Println("Broadcast", err)
		}
	}
}

// BroadcastError takes a JSON-RPC error and sends it to all connections
func (s *Server) BroadcastError(base *Base, response *Error) {
	for _, conn := range s.conns {
		if err := s.sendError(conn, base, response); err != nil {
			//log.Println("BroadcastError", err)
		}
	}
}

func (s *Server) startListen(listen net.Listener) {
	for {
		conn, err := listen.Accept()
		if err != nil {
			// aqui hay un error en potencia
			return
		}
		s.plug(conn)
		go (func(c net.Conn) {
			s.handleClient(c)
			s.unplug(c)
		})(conn)
	}
}

// Start prepares and launches the json-rpc server
func (s *Server) Start() (err error) {
	listen, err := net.Listen("tcp", s.Address)
	if err != nil {
		return err
	}

	go s.startListen(listen)

	s.plugBlocker = &sync.Mutex{}
	s.writeBlocker = &sync.Mutex{}
	s.conns = make([]net.Conn, 0)
	s.listen = listen
	s.registersSource = make(map[string]map[string]RemoteProcedure)
	s.registersTarget = make(map[string]map[string]DBRemoteProcedure)

	return nil
}
