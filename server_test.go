package jsonrpc

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"net"
	"sync"
	"testing"
)

func Test_Serve_Start_one_instance__OK(t *testing.T) {
	address := ":12345"
	done := make(chan bool)
	server := Server{Address: address}
	ready := make(chan bool)

	go (func() {
		err := server.Start(&ready)
		if err != nil {
			t.Error(err)
			done <- true
		}
	})()

	go (func() {
		<-ready

		_, err := net.Dial("tcp", address)
		if err != nil {
			t.Error(err)
		}

		done <- true
	})()

	<-done
}

func Test_Serve_Start_two_instances__fail(t *testing.T) {
	address := ":12346"
	done := make(chan bool)
	serverExtra := Server{Address: address}
	readyExtra := make(chan bool)

	server := Server{Address: address}
	ready := make(chan bool)

	go (func() {
		err := serverExtra.Start(&readyExtra)
		if err != nil {
			t.Error(err)
			done <- true
		}
	})()

	go (func() {
		<-readyExtra
		err := server.Start(&ready)
		if err == nil {
			t.Error("Expecting fail")
		}
		done <- true
	})()

	<-done
}

func Test_Serve_Start_then_Close__OK(t *testing.T) {
	address := ":12347"
	done := make(chan bool)
	server := Server{Address: address}
	ready := make(chan bool)

	go (func() {
		err := server.Start(&ready)
		if err != nil {
			t.Error(err)
			done <- true
		}
	})()

	go (func() {
		<-ready

		if _, err := net.Dial("tcp", address); err != nil {
			t.Error(err)
			done <- true
			return
		}

		server.Close()
		if _, err := net.Dial("tcp", address); err == nil {
			t.Error("Expecting to fail")
		}
		done <- true
	})()

	<-done
}

func Test_Serve_Send_Incorrect_JSON__fail(t *testing.T) {
	address := ":12347"
	done := make(chan bool)
	server := Server{Address: address}
	ready := make(chan bool)

	go (func() {
		err := server.Start(&ready)
		if err != nil {
			t.Error(err)
			done <- true
		}
	})()

	go (func() {
		<-ready
		conn, err := net.Dial("tcp", address)
		if err != nil {
			t.Error(err)
		}

		msg := []byte("!json\n")
		scanner := bufio.NewScanner(conn)
		conn.Write(msg)
		for scanner.Scan() {
			raw := scanner.Bytes()
			actual := string(raw)
			expected := `{"Code":-32700,"Message":"Parse error","Data":"invalid character '!' looking for beginning of value"}`
			if actual != expected {
				t.Errorf("\nexpect: %s\n!=\nactual: %s", expected, actual)
			}
			break
		}

		done <- true
	})()

	<-done
	server.Close()
}

func Test_Serve_Send_unknown_method_JSON__OK(t *testing.T) {
	address := ":12347"
	done := make(chan bool)
	server := Server{Address: address}
	ready := make(chan bool)

	go (func() {
		err := server.Start(&ready)
		if err != nil {
			t.Error(err)
			done <- true
		}
	})()

	go (func() {
		<-ready
		conn, err := net.Dial("tcp", address)
		if err != nil {
			t.Error(err)
		}

		request := Request{}
		request.ID = "ID"
		request.Method = "Unknown Method"
		request.Context = "Global"

		msg, _ := json.Marshal(request)

		scanner := bufio.NewScanner(conn)
		conn.Write(msg)
		conn.Write([]byte("\n"))
		for scanner.Scan() {
			raw := scanner.Bytes()
			actual := string(raw)
			expected := `{"Code":-32700,"Message":"Method not found","Data":{"ID":"ID","Method":"Unknown Method"}}`
			if actual != expected {
				t.Errorf("\nexpect: %s\n!=\nactual: %s", expected, actual)
			}
			break
		}

		done <- true
	})()

	<-done
	server.Close()
}

func Test_Serve_Send_incorrect_context__FAIL(t *testing.T) {
	address := ":12347"
	done := make(chan bool)
	server := Server{Address: address}
	ready := make(chan bool)

	go (func() {
		err := server.Start(&ready)
		if err != nil {
			t.Error(err)
			done <- true
		}
	})()

	go (func() {
		<-ready
		conn, err := net.Dial("tcp", address)
		if err != nil {
			t.Error(err)
		}

		request := Request{}
		request.ID = "ID"
		request.Method = "Unknown Method"
		request.Context = 434

		msg, _ := json.Marshal(request)

		scanner := bufio.NewScanner(conn)
		conn.Write(msg)
		conn.Write([]byte("\n"))
		for scanner.Scan() {
			raw := scanner.Bytes()
			actual := string(raw)
			expected := `{"Code":-32600,"Message":"Invalid Request","Data":{"Error":"Incorrect context 434","ID":"ID","Method":"Unknown Method"}}`
			if actual != expected {
				t.Errorf("\nexpect: %s\n!=\nactual: %s", expected, actual)
			}
			break
		}

		done <- true
	})()

	<-done
	server.Close()
}

func Test_Serve_Register_One_Ctx_One_Method__OK(t *testing.T) {
	address := ":12347"
	done := make(chan bool)
	server := Server{Address: address}
	ready := make(chan bool)

	pang :=
		func(request *Request) (result *interface{}) {
			var pong interface{} = "Pung"
			result = &pong
			return
		}

	go (func() {
		err := server.Start(&ready)
		if err != nil {
			t.Error(err)
			done <- true
		}
	})()

	go (func() {
		<-ready
		server.RegisterSource("Pang", "Global", pang)
		conn, err := net.Dial("tcp", address)
		if err != nil {
			t.Error(err)
		}

		request := Request{}
		request.ID = "ID"
		request.Method = "Pang"
		request.Context = "Global"

		msg, _ := json.Marshal(request)

		scanner := bufio.NewScanner(conn)
		conn.Write(msg)
		conn.Write([]byte("\n"))
		for scanner.Scan() {
			raw := scanner.Bytes()
			actual := string(raw)
			expected := `{"ID":"ID","Method":"Pang","Context":"Global","Result":"Pung","Error":null}`
			if actual != expected {
				t.Errorf("\nexpect: %s\n!=\nactual: %s", expected, actual)
			}
			break
		}

		done <- true
	})()

	<-done
	server.Close()
}

func Test_Serve_Register_One_Ctx_Two_Methods__OK(t *testing.T) {
	address := ":12347"
	done := make(chan bool)
	server := Server{Address: address}
	ready := make(chan bool)

	ping :=
		func(request *Request) (result *interface{}) {
			var pong interface{} = "Pong"
			result = &pong
			return
		}

	pang :=
		func(request *Request) (result *interface{}) {
			var pong interface{} = "Pung"
			result = &pong
			return
		}

	go (func() {
		err := server.Start(&ready)
		if err != nil {
			t.Error(err)
			done <- true
		}
	})()

	go (func() {
		<-ready
		server.RegisterSource("Ping", "Global", ping)
		server.RegisterSource("Pang", "Global", pang)
		conn, err := net.Dial("tcp", address)
		if err != nil {
			t.Error(err)
		}

		func() {
			request := Request{}
			request.ID = "ID"
			request.Method = "Ping"
			request.Context = "Global"

			msg, _ := json.Marshal(request)

			scanner := bufio.NewScanner(conn)
			conn.Write(msg)
			conn.Write([]byte("\n"))
			for scanner.Scan() {
				raw := scanner.Bytes()
				actual := string(raw)
				expected := `{"ID":"ID","Method":"Ping","Context":"Global","Result":"Pong","Error":null}`
				if actual != expected {
					t.Errorf("\nexpect: %s\n!=\nactual: %s", expected, actual)
				}
				break
			}
		}()

		func() {
			request := Request{}
			request.ID = "ID"
			request.Method = "Pang"
			request.Context = "Global"

			msg, _ := json.Marshal(request)

			scanner := bufio.NewScanner(conn)
			conn.Write(msg)
			conn.Write([]byte("\n"))
			for scanner.Scan() {
				raw := scanner.Bytes()
				actual := string(raw)
				expected := `{"ID":"ID","Method":"Pang","Context":"Global","Result":"Pung","Error":null}`
				if actual != expected {
					t.Errorf("\nexpect: %s\n!=\nactual: %s", expected, actual)
				}
				break
			}
		}()

		done <- true
	})()

	<-done
	server.Close()
}

func Test_Serve_Register_One_Complex_Ctx_One_Method__OK(t *testing.T) {
	address := ":12347"
	done := make(chan bool)
	server := Server{Address: address}
	ready := make(chan bool)

	pang :=
		func(request *Request) (result *interface{}) {
			var pong interface{} = "Pung"
			result = &pong
			return
		}

	go (func() {
		err := server.Start(&ready)
		if err != nil {
			t.Error(err)
			done <- true
		}
	})()

	go (func() {
		<-ready

		complexCtx := map[string]interface{}{"Source": "Global"}
		server.RegisterSource("Pang", "Global", pang)
		conn, err := net.Dial("tcp", address)
		if err != nil {
			t.Error(err)
		}

		request := Request{}
		request.ID = "ID"
		request.Method = "Pang"
		request.Context = complexCtx

		msg, _ := json.Marshal(request)

		scanner := bufio.NewScanner(conn)
		conn.Write(msg)
		conn.Write([]byte("\n"))
		for scanner.Scan() {
			raw := scanner.Bytes()
			actual := string(raw)
			expected := `{"ID":"ID","Method":"Pang","Context":{"Source":"Global"},"Result":"Pung","Error":null}`
			if actual != expected {
				t.Errorf("\nexpect: %s\n!=\nactual: %s", expected, actual)
			}
			break
		}

		done <- true
	})()

	<-done
	server.Close()
}

func Test_Serve_Register_One_Complex_Ctx_Two_Methods__OK(t *testing.T) {
	address := ":12347"
	done := make(chan bool)
	server := Server{Address: address}
	ready := make(chan bool)

	ping :=
		func(request *Request) (result *interface{}) {
			var pong interface{} = "Pong"
			result = &pong
			return
		}

	pang :=
		func(request *Request) (result *interface{}) {
			var pong interface{} = "Pung"
			result = &pong
			return
		}

	go (func() {
		err := server.Start(&ready)
		if err != nil {
			t.Error(err)
			done <- true
		}
	})()

	go (func() {
		<-ready

		complexCtx := map[string]interface{}{"Source": "Global"}
		server.RegisterSource("Ping", "Global", ping)
		server.RegisterSource("Pang", "Global", pang)
		conn, err := net.Dial("tcp", address)
		if err != nil {
			t.Error(err)
		}

		func() {
			request := Request{
				Base: Base{
					ID:      "ID",
					Method:  "Ping",
					Context: complexCtx,
				},
			}

			msg, _ := json.Marshal(request)

			scanner := bufio.NewScanner(conn)
			conn.Write(msg)
			conn.Write([]byte("\n"))
			for scanner.Scan() {
				raw := scanner.Bytes()
				actual := string(raw)
				expected := `{"ID":"ID","Method":"Ping","Context":{"Source":"Global"},"Result":"Pong","Error":null}`
				if actual != expected {
					t.Errorf("\nexpect: %s\n!=\nactual: %s", expected, actual)
				}
				break
			}
		}()

		func() {
			request := Request{
				Base: Base{
					ID:      "ID",
					Method:  "Pang",
					Context: complexCtx,
				},
			}

			msg, _ := json.Marshal(request)

			scanner := bufio.NewScanner(conn)
			conn.Write(msg)
			conn.Write([]byte("\n"))
			for scanner.Scan() {
				raw := scanner.Bytes()
				actual := string(raw)
				expected := `{"ID":"ID","Method":"Pang","Context":{"Source":"Global"},"Result":"Pung","Error":null}`
				if actual != expected {
					t.Errorf("\nexpect: %s\n!=\nactual: %s", expected, actual)
				}
				break
			}
		}()

		done <- true
	})()

	<-done
	server.Close()
}

func Test_Serve_connect_disconnect_two_clients__OK(t *testing.T) {
	address := ":12347"
	done := make(chan bool)
	server := Server{Address: address}
	ready := make(chan bool)

	go (func() {
		err := server.Start(&ready)
		if err != nil {
			t.Error(err)
			done <- true
		}
	})()

	go (func() {
		<-ready

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

		done <- true
	})()

	<-done
	server.Close()
}

func Test_Serve_Register_One_Complex_Ctx_One_Method_ProcessNotification__OK(t *testing.T) {
	address := ":12347"
	var w sync.WaitGroup
	done := make(chan bool)
	server := Server{Address: address}
	ready := make(chan bool)

	var ping DBRemoteProcedure = func(db *sql.DB) RemoteProcedure {
		return func(request *Request) (result *interface{}) {
			var pong interface{} = "Pong"
			result = &pong
			return
		}
	}

	go (func() {
		err := server.Start(&ready)
		if err != nil {
			t.Error(err)
			done <- true
		}
	})()

	go (func() {
		<-ready

		conn1, err := net.Dial("tcp", address)
		w.Add(1)
		if err != nil {
			t.Error(err)
		}

		conn2, err := net.Dial("tcp", address)
		w.Add(1)
		if err != nil {
			t.Error(err)
		}

		conn3, err := net.Dial("tcp", address)
		w.Add(1)
		if err != nil {
			t.Error(err)
		}

		complexCtx := map[string]interface{}{"Target": "Global"}
		server.RegisterTarget("Ping", "Global", ping)

		request := Request{
			Base: Base{
				ID:      "ID",
				Method:  "Ping",
				Context: complexCtx,
			},
		}

		server.ProcessNotification(&request, nil)

		func() {
			scanner := bufio.NewScanner(conn1)
			for scanner.Scan() {
				raw := scanner.Bytes()
				actual := string(raw)
				expected := `{"ID":"ID","Method":"Ping","Context":{"Target":"Global"},"Result":"Pong","Error":null}`
				w.Done()
				if actual != expected {
					t.Errorf("\nexpect: %s\n!=\nactual: %s", expected, actual)
				}
				break
			}
		}()

		func() {
			scanner := bufio.NewScanner(conn2)
			for scanner.Scan() {
				raw := scanner.Bytes()
				actual := string(raw)
				expected := `{"ID":"ID","Method":"Ping","Context":{"Target":"Global"},"Result":"Pong","Error":null}`
				w.Done()
				if actual != expected {
					t.Errorf("\nexpect: %s\n!=\nactual: %s", expected, actual)
				}
				break
			}
		}()

		func() {
			scanner := bufio.NewScanner(conn3)
			for scanner.Scan() {
				raw := scanner.Bytes()
				actual := string(raw)
				expected := `{"ID":"ID","Method":"Ping","Context":{"Target":"Global"},"Result":"Pong","Error":null}`
				w.Done()
				if actual != expected {
					t.Errorf("\nexpect: %s\n!=\nactual: %s", expected, actual)
				}
				break
			}
		}()

		w.Wait()
		done <- true
	})()

	<-done
	server.Close()
}
