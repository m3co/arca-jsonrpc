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
	base := &Base{
		ID:      request.ID,
		Method:  request.Method,
		Context: request.Context,
	}

	ctx, err := getFieldFromContext("Target", request.Context)
	if err != nil {
		s.BroadcastError(base, &Error{
			Message: "Internal error",
			Code:    -32603,
			Data: map[string]string{
				"Error":  fmt.Sprint(err),
				"Method": request.Method,
				"ID":     request.ID,
			},
		})
	}

	response, err := s.findAndExecuteHandlerInTarget(ctx, request, base, db)
	if err != nil {
		s.BroadcastError(base, &Error{
			Message: "Internal error",
			Code:    -32603,
			Data: map[string]string{
				"Error":  fmt.Sprint(err),
				"Method": request.Method,
				"ID":     request.ID,
			},
		})
	}
	if response != nil {
		msg, _ := json.Marshal(response)
		s.Broadcast(msg)
	}
}

// ProcessRequest takes a request and a conn, and depending on the request it
// matches a handler, calls that handler with the request as parametr and that
// result sends it through the given conn
func (s *Server) ProcessRequest(
	request *Request, conn net.Conn) {

	base := &Base{
		ID:      request.ID,
		Method:  request.Method,
		Context: request.Context,
	}

	src := "Source"
	if conn == nil {
		src = "Target"
	}
	ctx, err := getFieldFromContext(src, request.Context)
	if err != nil {
		//log.Println("ProcessRequest:getFieldFromContext", err)
		if err1 := s.sendError(conn, base, &Error{
			Message: "Invalid Request",
			Code:    -32600,
			Data: map[string]string{
				"Method": request.Method,
				"ID":     request.ID,
				"Error":  fmt.Sprint(err),
			},
		}); err1 != nil {
			//log.Println("ProcessRequest:getFieldFromContext:sendError", err)
		}
		return
	}

	/*
		if request.Params == nil {
			if err1 := s.sendError(conn, base, &Error{
				Message: "Invalid params",
				Code:    -32602,
				Data: map[string]string{
					"Method": request.Method,
					"ID":     request.ID,
					"Error":  "Params in request not found",
				},
			}); err1 != nil {
				//log.Println("ProcessRequest:getFieldFromContext:Params", err)
			}
			return
		}
	*/

	response, err := s.findAndExecuteHandlerInSource(ctx, request, base)
	if err != nil {
		//log.Println("ProcessRequest:findAndExecuteHandlerInSource", err)
		if err == errMethodNotMatch {
			if err := s.sendError(conn, base, &Error{
				Message: "Method not found",
				Code:    -32601,
				Data: map[string]string{
					"Method": request.Method,
					"ID":     request.ID,
				},
			}); err != nil {
				//log.Println("ProcessRequest:findAndExecuteHandlerInSource:errMethodNotMatch:sendError", err)
			}
		} else {
			if err := s.sendError(conn, base, &Error{
				Message: "Internal error",
				Code:    -32603,
				Data: map[string]string{
					"Error":  fmt.Sprint(err),
					"Method": request.Method,
					"ID":     request.ID,
				},
			}); err != nil {
				//log.Println("ProcessRequest:findAndExecuteHandlerInSource:sendError", err)
			}
		}
	} else if response != nil {
		if err := s.send(conn, response); err != nil {
			//log.Println("ProcessRequest:response:send", err)
		}
	}
}
