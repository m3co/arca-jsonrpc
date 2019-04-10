package jsonrpc

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
)

// ProcessNotification whatever
func (s *Server) ProcessNotification(
	request *Request, db *sql.DB) {
	base := Base{
		ID:      request.ID,
		Method:  request.Method,
		Context: request.Context,
	}

	ctx, err := getFieldFromContext("Target", request.Context)
	if err != nil {
		return
	}

	response, err := s.findAndExecuteHandlerInTarget(ctx, request, &base, db)
	if response != nil {
		msg, _ := json.Marshal(response)
		s.Broadcast(msg)
	}
}

// ProcessRequest takes a request and a conn, and depending on the request it
// matches a handler, calls that handler with the request as parametr and that
// result sends it through the given conn
func (s *Server) ProcessRequest(
	request *Request, conn *net.Conn) {

	base := Base{
		ID:      request.ID,
		Method:  request.Method,
		Context: request.Context,
	}

	ctx, err := getFieldFromContext("Source", request.Context)
	if err != nil {
		(*s).sendError(conn, &Error{
			Message: "Invalid Request",
			Code:    -32600,
			Data: map[string]string{
				"Method": request.Method,
				"ID":     request.ID,
				"Error":  fmt.Sprint(err),
			},
		})
		return
	}

	response, err := s.findAndExecuteHandlerInSource(ctx, request, &base)
	if err != nil {
		(*s).sendError(conn, &Error{
			Message: "Method not found",
			Code:    -32700,
			Data: map[string]string{
				"Method": request.Method,
				"ID":     request.ID,
			},
		})
	} else if response != nil {
		(*s).send(conn, response)
	}
}
