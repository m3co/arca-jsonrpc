package jsonrpc

import (
	"database/sql"
	"fmt"
	"log"
	"net"
)

// ProcessNotification whatever
func (s *Server) ProcessNotification(
	request *Request, db *sql.DB) (interface{}, error) {
	base := Base{
		ID:      request.ID,
		Method:  request.Method,
		Context: request.Context,
	}

	ctx, err := getFieldFromContext("Target", request.Context)
	if err != nil {
		return nil, err
	}

	response, err := s.findAndExecuteHandlerInTarget(ctx, request, &base, db)
	if err != nil {
		return &Error{
			Message: "Internal error",
			Code:    -32603,
			Data: map[string]string{
				"Error":  fmt.Sprint(err),
				"Method": request.Method,
				"ID":     request.ID,
			},
		}, err
	}
	return response, err
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
		log.Println("ProcessRequest", err)
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
		log.Println("ProcessRequest", err)
		if err == errMethodNotMatch {
			(*s).sendError(conn, &Error{
				Message: "Method not found",
				Code:    -32700,
				Data: map[string]string{
					"Method": request.Method,
					"ID":     request.ID,
				},
			})
		} else {
			(*s).sendError(conn, &Error{
				Message: "Internal error",
				Code:    -32603,
				Data: map[string]string{
					"Error":  fmt.Sprint(err),
					"Method": request.Method,
					"ID":     request.ID,
				},
			})
		}
	} else if response != nil {
		(*s).send(conn, response)
	}
}
