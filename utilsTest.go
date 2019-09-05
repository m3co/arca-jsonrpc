package jsonrpc

import (
	"bufio"
	"encoding/json"
	"net"
	"testing"
)

var address = ":12345"

func startServer() (
	server *Server, errch error) {
	return startServerWithAddress(address)
}

func startServerWithAddress(address string) (
	server *Server, err error) {
	server = &Server{Address: address}
	err = server.Start()
	return
}

func startServerAndClient(t *testing.T) (
	server *Server, conn net.Conn, err error) {
	server, err = startServer()
	if err != nil {
		t.Error(err)
		return
	}

	conn, err = net.Dial("tcp", address)
	if err != nil {
		t.Error(err)
		server.Close()
	}
	return
}

func sendJSONAndReceive(conn *net.Conn, request *Request) string {
	msg, _ := json.Marshal(request)
	return sendAndReceive(conn, msg)
}

func sendAndReceive(conn *net.Conn, request []byte) (response string) {
	scanner := bufio.NewScanner(*conn)
	(*conn).Write(request)
	(*conn).Write([]byte("\n"))
	for scanner.Scan() {
		raw := scanner.Bytes()
		response = string(raw)
		break
	}
	return
}

func assertExpectedVsActualAndClose(t *testing.T, expected, actual string, server *Server) {
	if expected != actual {
		t.Errorf("\nexpect %s\nactual %s", expected, actual)
	}
	if server != nil {
		server.Close()
	}
}
