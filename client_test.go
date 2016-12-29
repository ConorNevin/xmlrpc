package xmlrpc

import (
	"testing"
	"time"
	"net"
	"strings"
	"errors"
	"net/http"
	"github.com/divan/gorilla-xmlrpc/xml"
	"github.com/gorilla/rpc"
)

var (
	client *Client
	listener net.Listener
)

func TestClient(t *testing.T ) {
	client, listener = newClient(t)

	t.Run("CallWithoutArgs", CallWithoutArgs)
	t.Run("CallWithOneArg", CallWithOneArg)
	t.Run("CallWithTwoArgs", CallWithTwoArgs)
	t.Run("TwoCalls", TwoCalls)
	t.Run("FailedCall", FailedCall)
}

func CallWithoutArgs(t *testing.T) {
	var result time.Time
	if err := client.Call("TestServer.Time", nil, &result); err != nil {
		t.Fatalf("TestServer.Time call error: %v", err)
	}
}

func CallWithOneArg(t *testing.T) {
	var result string
	if err := client.Call("TestServer.Upcase", "xmlrpc", &result); err != nil {
		t.Fatalf("TestServer.Upcase call error: %v", err)
	}

	if result != "XMLRPC" {
		t.Fatalf("Unexpected result of service.Upcase: %s != %s", "XMLRPC", result)
	}
}

func CallWithTwoArgs(t *testing.T) {
	var sum int
	if err := client.Call("TestServer.Sum", []int{2, 3}, &sum); err != nil {
		t.Fatalf("TestServer.Sum call error: %v", err)
	}

	if sum != 5 {
		t.Fatalf("Unexpected result of service.sum: %d != %d", 5, sum)
	}
}

func TwoCalls(t *testing.T) {
	var upcase string
	if err := client.Call("TestServer.Upcase", "xmlrpc", &upcase); err != nil {
		t.Fatalf("TestServer.Upcase call error: %v", err)
	}

	var sum int
	if err := client.Call("TestServer.Sum", []int{2, 3}, &sum); err != nil {
		t.Fatalf("TestServer.Sum call error: %v", err)
	}

}

func FailedCall(t *testing.T) {
	var result int
	if err := client.Call("TestServer.Error", nil, &result); err == nil {
		t.Fatal("expected TestServer.Error returns error, but it didn't")
	}
}

func newClient(t *testing.T) (*Client, net.Listener) {
	server := rpc.NewServer()

	xmlrpcCodec := xml.NewCodec()
	rpcServer := new(TestServer)

	server.RegisterCodec(xmlrpcCodec, "text/xml")
	server.RegisterService(rpcServer, "TestServer")
	http.Handle("/", server)

	listener, err := net.Listen("tcp", ":5001")
	if err != nil {
		t.Fatalf("Can't create test RPC server: %v", err)
	}

	go http.Serve(listener, nil)
	
	client, err := NewClient("http://localhost:5001/", nil)
	if err != nil {
		t.Fatalf("Can't create client: %v", err)
	}

	return client, listener
}

type TestServer struct{}

func (s *TestServer) Time(req *http.Request, arg *struct{}, result *struct{Out time.Time}) error {
	result.Out = time.Now()
	return nil
}

func (s *TestServer) Upcase(req *http.Request, arg *struct{In string}, result *struct{Out string}) error {
	result.Out = strings.ToUpper(arg.In)
	return nil
}

func (s *TestServer) Sum(req *http.Request, arg *struct{In []int}, result *struct{Out int}) error {
	if len(arg.In) != 2 {
		return errors.New("You can only sum two elements")
	}
	result.Out = arg.In[0] + arg.In[1]
	return nil
}

func (s *TestServer) Error(req *http.Request, arg *struct{}, result *struct{}) error {
	return errors.New("Server error")
}