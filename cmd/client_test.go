package cmd

import (
	"context"
	"testing"

	"github.com/gogo/protobuf/types"
	"go.bryk.io/x/net/rpc"
	samplev1 "go.bryk.io/x/net/rpc/sample/v1"
	"google.golang.org/grpc"
)

func TestClient(t *testing.T) {
	// Start test server
	srvOptions := []rpc.ServerOption{
		rpc.WithNetworkInterface(rpc.NetworkInterfaceAll),
		rpc.WithPort(8989),
		rpc.WithLogger(rpc.LoggingOptions{}),
		rpc.WithPanicRecovery(),
		rpc.WithService(&rpc.Service{
			Setup: func(server *grpc.Server) {
				samplev1.RegisterEchoAPIServer(server, &samplev1.EchoHandler{})
			},
		}),
	}
	srv, err := rpc.NewServer(srvOptions...)
	if err != nil {
		t.Fatal(err)
	}
	ready := make(chan bool)
	go func() {
		_ = srv.Start(ready)
	}()
	<-ready

	conn, err := rpc.NewClientConnection("127.0.0.1:8989", rpc.WithClientLogger(rpc.LoggingOptions{}))
	if err != nil {
		t.Error(err)
	}
	cl := samplev1.NewEchoAPIClient(conn)

	t.Run("Ping", func(t *testing.T) {
		if _, err := cl.Ping(context.TODO(), &types.Empty{}); err != nil {
			t.Error(err)
		}
	})

	t.Run("EchoRequest", func(t *testing.T) {
		req := &samplev1.EchoRequest{Value:"hello world"}
		res, err := cl.Request(context.TODO(), req)
		if err != nil {
			t.Error(err)
		}
		if res.Result != "you said: hello world" {
			t.Error("invalid result")
		}
	})

	// Stop server
	defer func() {
		_ = srv.Stop()
	}()
}
