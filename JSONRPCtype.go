package jsonrpc

import (
	"database/sql"
	"net"
	"sync"
)

// Base is the base for both request and response structures
type Base struct {
	ID      string
	Method  string
	Context interface{}
}

// Error is the structure of JSON-RPC response
type Error struct {
	Code    int
	Message string
	Data    interface{}
}

// Request is the structure of JSON-RPC request
type Request struct {
	Base
	Params interface{}
}

// Response is the structure of JSON-RPC response
type Response struct {
	Base
	Result interface{}
	Error  interface{}
}

// RemoteProcedure represents the function-handler that matches a given request
//   TODO: allows to return an error!
type RemoteProcedure func(request *Request) (result interface{}, err error)

// DBRemoteProcedure represents the function-handler that goes through the
// db pool and executes the RemoteProcedure
type DBRemoteProcedure func(db *sql.DB) RemoteProcedure

// Server represents the arca-jsonrpc server
type Server struct {
	Address         string
	plugBlocker     *sync.Mutex
	writeBlocker    *sync.Mutex
	closeBlocker    *sync.Mutex
	conns           []*net.Conn
	listen          *net.Listener
	registersSource map[string]map[string]RemoteProcedure
	registersTarget map[string]map[string]DBRemoteProcedure
}
