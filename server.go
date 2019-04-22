package jsonrpc

import (
	"net"
	"sync"
)

// Close takes the listen and close channel and closes them
func (s *Server) Close() {
	(*s.listen).Close()
	s.close <- true
}

// Broadcast sends to all the active connections the given message
func (s *Server) Broadcast(msg []byte) {
	for _, conn := range s.conns {
		s.write(conn, msg)
	}
}

// Start prepares and launches the json-rpc server
func (s *Server) Start(ready *chan bool) (err error) {
	listen, err := net.Listen("tcp", s.Address)
	if err != nil {
		return err
	}

	s.blocker = &sync.Mutex{}
	s.close = make(chan bool)
	s.conns = make([]*net.Conn, 0)
	s.listen = &listen
	s.registersSource = make(map[string]map[string]RemoteProcedure)
	s.registersTarget = make(map[string]map[string]DBRemoteProcedure)

	go (func() {
		var wait sync.WaitGroup
		for {
			conn, err := listen.Accept()
			if err != nil {
				return
			}
			wait.Wait()
			wait.Add(1)
			s.plug(&conn)
			wait.Done()
			go (func(conn *net.Conn) {
				s.handleClient(conn)
				wait.Wait()
				wait.Add(1)
				s.unplug(conn)
				wait.Done()
			})(&conn)
		}
	})()

	*ready <- true
	<-s.close
	defer s.Close()
	return nil
}
