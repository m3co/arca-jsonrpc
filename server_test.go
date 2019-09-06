package jsonrpc

import (
	"database/sql"
	"fmt"
	"net"
	"testing"
)

func Test_Serve_Start_one_instance__OK(t *testing.T) {
	server, _, _ := startServerAndClient(t)
	server.Close()
}

func Test_Serve_Start_two_instances__fail(t *testing.T) {
	server, errServer := startServer()
	if errServer != nil {
		t.Error(errServer)
		return
	}

	_, err := startServer()
	if err.Error() != fmt.Sprintf(
		"listen tcp %s: bind: address already in use", address) {
		t.Error(err)
	}

	server.Close()
}

func Test_Serve_Start_then_Close__OK(t *testing.T) {
	server, _, _ := startServerAndClient(t)

	server.Close()
	if _, err := net.Dial("tcp", address); err == nil {
		t.Error("Expecting error")
	}
}

func Test_Serve_Send_Incorrect_JSON__fail(t *testing.T) {
	server, conn, err := startServerAndClient(t)
	if err != nil {
		return
	}

	expected := `{"ID":"","Method":"","Context":null,"Result":null,"Error":{"Code":-32700,"Message":"Parse error","Data":"invalid character '!' looking for beginning of value"}}`
	actual := sendAndReceive(&conn, []byte("!json"))
	assertExpectedVsActualAndClose(t, expected, actual, server)
}

func Test_Serve_Send_unknown_method_JSON__OK(t *testing.T) {
	server, conn, err := startServerAndClient(t)
	if err != nil {
		return
	}

	request := Request{}
	request.ID = "ID"
	request.Method = "Unknown Method"
	request.Context = "Global"

	expected := `{"ID":"ID","Method":"Unknown Method","Context":"Global","Result":null,"Error":{"Code":-32700,"Message":"Method not found","Data":{"ID":"ID","Method":"Unknown Method"}}}`
	actual := sendJSONAndReceive(&conn, &request)
	assertExpectedVsActualAndClose(t, expected, actual, server)
}

func Test_Serve_Send_incorrect_context__FAIL(t *testing.T) {
	server, conn, err := startServerAndClient(t)
	if err != nil {
		return
	}

	request := Request{}
	request.ID = "ID"
	request.Method = "Unknown Method"
	request.Context = 434

	expected := `{"ID":"ID","Method":"Unknown Method","Context":434,"Result":null,"Error":{"Code":-32600,"Message":"Invalid Request","Data":{"Error":"Incorrect context 434","ID":"ID","Method":"Unknown Method"}}}`
	actual := sendJSONAndReceive(&conn, &request)
	assertExpectedVsActualAndClose(t, expected, actual, server)
}

func Test_Serve_Register_One_Ctx_One_Method__OK(t *testing.T) {
	server, conn, err := startServerAndClient(t)
	if err != nil {
		return
	}

	pung :=
		func(request *Request) (result interface{}, err error) {
			var pong interface{} = "Pung"
			result = &pong
			return
		}

	server.RegisterSource("Pung", "Global", pung)
	request := Request{}
	request.ID = "ID"
	request.Method = "Pung"
	request.Context = "Global"

	expected := `{"ID":"ID","Method":"Pung","Context":"Global","Result":"Pung","Error":null}`
	actual := sendJSONAndReceive(&conn, &request)
	assertExpectedVsActualAndClose(t, expected, actual, server)
}

func Test_Serve_Register_One_Ctx_Two_Methods__OK(t *testing.T) {
	server, conn, err := startServerAndClient(t)
	if err != nil {
		return
	}

	ping :=
		func(request *Request) (result interface{}, err error) {
			var pong interface{} = "Pong"
			result = &pong
			return
		}

	pang :=
		func(request *Request) (result interface{}, err error) {
			var pong interface{} = "Pung"
			result = &pong
			return
		}

	server.RegisterSource("Ping", "Global", ping)
	server.RegisterSource("Pang", "Global", pang)

	var request Request
	request = Request{}
	request.ID = "ID"
	request.Context = "Global"

	request.Method = "Ping"
	expectedPing := `{"ID":"ID","Method":"Ping","Context":"Global","Result":"Pong","Error":null}`
	actualPing := sendJSONAndReceive(&conn, &request)
	assertExpectedVsActualAndClose(t, expectedPing, actualPing, nil)

	request.Method = "Pang"
	expectedPang := `{"ID":"ID","Method":"Pang","Context":"Global","Result":"Pung","Error":null}`
	actualPang := sendJSONAndReceive(&conn, &request)
	assertExpectedVsActualAndClose(t, expectedPang, actualPang, server)
}

func Test_Serve_Register_One_Complex_Ctx_One_Method__OK(t *testing.T) {
	server, conn, err := startServerAndClient(t)
	if err != nil {
		return
	}

	pang :=
		func(request *Request) (result interface{}, err error) {
			var pong interface{} = "Pung"
			result = &pong
			return
		}
	complexCtx := map[string]interface{}{"Source": "Global"}
	server.RegisterSource("Pang", "Global", pang)

	request := Request{}
	request.ID = "ID"
	request.Method = "Pang"
	request.Context = complexCtx

	expected := `{"ID":"ID","Method":"Pang","Context":{"Source":"Global"},"Result":"Pung","Error":null}`

	actual := sendJSONAndReceive(&conn, &request)
	assertExpectedVsActualAndClose(t, expected, actual, server)
}

func Test_Serve_Register_One_Complex_Ctx_Two_Methods__OK(t *testing.T) {
	server, conn, err := startServerAndClient(t)
	if err != nil {
		return
	}

	ping :=
		func(request *Request) (result interface{}, err error) {
			var pong interface{} = "Pong"
			result = &pong
			return
		}

	pang :=
		func(request *Request) (result interface{}, err error) {
			var pong interface{} = "Pung"
			result = &pong
			return
		}

	complexCtx := map[string]interface{}{"Source": "Global"}
	server.RegisterSource("Ping", "Global", ping)
	server.RegisterSource("Pang", "Global", pang)

	var request Request
	var expected, actual string
	request = Request{}
	request.ID = "ID"
	request.Context = complexCtx

	request.Method = "Ping"
	expected = `{"ID":"ID","Method":"Ping","Context":{"Source":"Global"},"Result":"Pong","Error":null}`
	actual = sendJSONAndReceive(&conn, &request)
	assertExpectedVsActualAndClose(t, expected, actual, nil)

	request.Method = "Pang"
	expected = `{"ID":"ID","Method":"Pang","Context":{"Source":"Global"},"Result":"Pung","Error":null}`
	actual = sendJSONAndReceive(&conn, &request)
	assertExpectedVsActualAndClose(t, expected, actual, server)

}

func Test_Serve_connect_disconnect_two_clients__OK(t *testing.T) {
	server, errServer := startServer()
	if errServer != nil {
		t.Error(errServer)
		return
	}

	conn1, err := net.Dial("tcp", address)
	if err != nil {
		t.Error(err)
	}

	conn2, err := net.Dial("tcp", address)
	if err != nil {
		t.Error(err)
	}

	conn2.Close()
	conn1.Close()

	server.Close()
}
func Test_Serve_Register_One_Complex_Ctx_One_Method_ProcessNotification__MethodNotFound(t *testing.T) {
	server, errServer := startServer()
	if errServer != nil {
		t.Error(errServer)
		return
	}

	ping :=
		func(db *sql.DB) RemoteProcedure {
			return func(request *Request) (result interface{}, err error) {
				var pong interface{} = "Pong"
				result = &pong
				return
			}
		}

	conn1, err := net.Dial("tcp", address)
	if err != nil {
		t.Error(err)
	}

	conn2, err := net.Dial("tcp", address)
	if err != nil {
		t.Error(err)
	}

	conn3, err := net.Dial("tcp", address)
	if err != nil {
		t.Error(err)
	}

	complexCtx := map[string]interface{}{"Target": "Global"}
	server.RegisterTarget("Ping", "Global", ping)

	request := Request{
		Base: Base{
			ID:      "ID",
			Method:  "Unknown",
			Context: complexCtx,
		},
	}

	server.ProcessNotification(&request, nil)

	response1 := receiveString(&conn1)
	response2 := receiveString(&conn2)
	response3 := receiveString(&conn3)

	expected := `{"ID":"ID","Method":"Unknown","Context":{"Target":"Global"},"Result":null,"Error":{"Code":-32603,"Message":"Internal error","Data":{"Error":"Method not found","ID":"ID","Method":"Unknown"}}}`
	assertExpectedVsActualAndClose(t, expected, response1, nil)
	assertExpectedVsActualAndClose(t, expected, response2, nil)
	assertExpectedVsActualAndClose(t, expected, response3, nil)

	server.Close()
}
