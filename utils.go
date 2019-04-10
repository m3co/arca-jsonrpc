package jsonrpc

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net"
)

// write sends the given message thorugh the given conn
func write(conn *net.Conn, msg []byte) {
	(*conn).Write(msg)
	(*conn).Write([]byte("\n"))
}

// send takes a JSON-RPC response and sends it thorugh the given conn
func (s *Server) send(conn *net.Conn, response *Response) {
	msg, _ := json.Marshal(response)
	write(conn, msg)
}

// send takes a JSON-RPC error and sends it thorugh the given conn
func (s *Server) sendError(conn *net.Conn, response *Error) {
	msg, _ := json.Marshal(response)
	write(conn, msg)
}

// plug appends a conn in the array of connections. Necessary for broadcasting
func (s *Server) plug(conn *net.Conn) {
	s.conns = append(s.conns, conn)
}

// unplug drops a conn in the array of connections. Necessary for broadcasting
func (s *Server) unplug(conn *net.Conn) {
	for i, value := range s.conns {
		if value == conn {
			s.conns[i] = s.conns[len(s.conns)-1]
			s.conns = s.conns[:len(s.conns)-1]
			return
		}
	}
}

// getFieldFromContext extracts from the context the value of the given field
func getFieldFromContext(
	field string, context interface{}) (ctx string, err error) {
	switch context.(type) {
	case map[string]interface{}:
		ctx = context.(map[string]interface{})[field].(string)
	case string:
		ctx = context.(string)
	default:
		err = fmt.Errorf("Incorrect context %v", context)
	}
	return
}

// findAndExecuteHandlerInTarget finds and executes the respective handler
// that fits the JSON-RP request based on the given context string ctx
func (s *Server) findAndExecuteHandlerInTarget(
	ctx string,
	request *Request, base *Base, db *sql.DB) (*Response, error) {

	if s.registersTarget[ctx] != nil {
		found := s.registersTarget[ctx][request.Method]
		if found != nil {
			result := found(db)(request)

			if result != nil {
				response := Response{
					Base:   *base,
					Result: result,
				}
				return &response, nil
			}
			return nil, nil
		}
	}
	return nil, errors.New("not match")
}

// findAndExecuteHandlerInSource finds and executes the respective handler
// that fits the JSON-RP request based on the given context string ctx
func (s *Server) findAndExecuteHandlerInSource(
	ctx string,
	request *Request, base *Base) (*Response, error) {

	if s.registersSource[ctx] != nil {
		found := s.registersSource[ctx][request.Method]
		if found != nil {
			result := found(request)

			if result != nil {
				response := Response{
					Base:   *base,
					Result: result,
				}
				return &response, nil
			}
			return nil, nil
		}
	}
	return nil, errors.New("not match")
}
