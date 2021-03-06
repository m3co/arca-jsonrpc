package jsonrpc

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net"
)

var (
	errMethodNotMatch  = errors.New("Method not found")
	errConnNilWhenSend = errors.New("Cannot send response if conn is nil")
)

// write sends the given message thorugh the given conn
func (s *Server) write(conn net.Conn, msg []byte) error {
	s.writeBlocker.Lock()
	defer s.writeBlocker.Unlock()
	if _, err := conn.Write(msg); err != nil {
		return err
	}
	if _, err := conn.Write([]byte("\n")); err != nil {
		return err
	}
	return nil
}

// send takes a JSON-RPC response and sends it thorugh the given conn
func (s *Server) send(conn net.Conn, response *Response) error {
	var err error
	var msg []byte
	if conn == nil {
		return errConnNilWhenSend
	}
	msg, err = json.Marshal(response)
	if err != nil {
		return err
	}
	return s.write(conn, msg)
}

// send takes a JSON-RPC error and sends it thorugh the given conn
func (s *Server) sendError(conn net.Conn, base *Base, err *Error) error {
	var err1 error
	var msg []byte
	if conn == nil {
		return errConnNilWhenSend
	}
	response := &Response{
		Base:  *base,
		Error: err,
	}
	msg, err1 = json.Marshal(response)
	if err1 != nil {
		return err1
	}
	return s.write(conn, msg)
}

// plug appends a conn in the array of connections. Necessary for broadcasting
func (s *Server) plug(conn net.Conn) {
	s.plugBlocker.Lock()
	defer s.plugBlocker.Unlock()
	s.conns = append(s.conns, conn)
}

// unplug drops a conn in the array of connections. Necessary for broadcasting
func (s *Server) unplug(conn net.Conn) {
	s.plugBlocker.Lock()
	defer s.plugBlocker.Unlock()
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
		value := context.(map[string]interface{})[field]
		if value != nil {
			ctx = value.(string)
		} else {
			err = fmt.Errorf("Incorrect context %v", context)
		}
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
	var result interface{}
	var err error

	if s.registersTarget[ctx] != nil {
		found := s.registersTarget[ctx][request.Method]

		if found != nil {
			func() {
				defer func() {
					if r := recover(); r != nil {
						err = fmt.Errorf("%v", r)
					}
				}()
				result, err = found(db)(request)
			}()

			if result != nil {
				response := Response{
					Base:   *base,
					Result: result,
				}
				return &response, err
			}
			return nil, err
		}
	}
	return nil, errMethodNotMatch
}

// findAndExecuteHandlerInSource finds and executes the respective handler
// that fits the JSON-RP request based on the given context string ctx
func (s *Server) findAndExecuteHandlerInSource(
	ctx string,
	request *Request, base *Base) (*Response, error) {
	var result interface{}
	var err error

	if s.registersSource[ctx] != nil {
		found := s.registersSource[ctx][request.Method]

		if found != nil {
			func() {
				defer func() {
					if r := recover(); r != nil {
						err = fmt.Errorf("%v", r)
					}
				}()
				result, err = found(request)
			}()

			if result != nil {
				response := Response{
					Base:   *base,
					Result: result,
				}
				return &response, err
			}
			return nil, err
		}
	}
	return nil, errMethodNotMatch
}
