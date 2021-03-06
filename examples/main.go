package main

import (
	"context"
	"flag"
	"fmt"

	daprClient "github.com/dapr/go-sdk/client"
	daprServer "github.com/dapr/go-sdk/service/grpc"
	"github.com/purefun/protoc-gen-dapr/examples/echo"
)

// dapr run -a echo_server -p 3000 -P grpc -- go run main.go --server
// dapr run -a echo_client -p 3000 -P grpc -- go run main.go --client

func main() {

	runClient := flag.Bool("client", false, "run client")
	runServer := flag.Bool("server", false, "run server")

	flag.Parse()

	if *runClient {
		client()
	}

	if *runServer {
		server()
	}
}

func client() {
	c, _ := daprClient.NewClient()
	echoClient := echo.NewEchoClient(c, "echo_server")

	req := new(echo.EchoRequest)
	req.Message = "Hey"

	resp, err := echoClient.Echo(context.Background(), req)

	if err != nil {
		fmt.Println("call echo error: ", err)
	} else {
		fmt.Println(resp.Message)
	}
}

type Handlers struct {
	echo.UnimplementedEchoServer
}

func (h *Handlers) Echo(ctx context.Context, in *echo.EchoRequest) (*echo.EchoResponse, error) {
	resp := new(echo.EchoResponse)
	resp.Message = in.GetMessage() + ", yo"
	return resp, nil
}

func (h *Handlers) mustEmbedUnimplementedEchoServer() {
	panic("not implemented") // TODO: Implement
}

func server() {
	s, err := daprServer.NewService(":3000")

	if err != nil {
		panic(err)
	}

	h := new(Handlers)

	echo.Register(s, h)

	s.Start()
}
