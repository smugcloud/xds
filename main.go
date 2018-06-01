package main

import (
	"context"
	"io"
	"log"
	"net"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"google.golang.org/grpc"
)

const (
	proto   = "tcp"
	iface   = "localhost"
	mgrPort = "19000"
)

type listenerServer struct {
}

func main() {
	lis, err := net.Listen(proto, iface+":"+mgrPort)
	if err != nil {
		log.Fatalf("Error getting listener: %v", err)
	}
	// var srv v2.ListenerDiscoveryServiceServer
	grpc := grpc.NewServer()
	v2.RegisterListenerDiscoveryServiceServer(grpc, &listenerServer{})
	log.Printf("Listening on: %v", "http://"+iface+":"+mgrPort)
	grpc.Serve(lis)
}

func (l *listenerServer) StreamListeners(sls v2.ListenerDiscoveryService_StreamListenersServer) error {
	log.Print("In the stream")
	for {
		dr, err := sls.Recv()
		log.Print(dr)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func (l *listenerServer) FetchListeners(context.Context, *v2.DiscoveryRequest) (*v2.DiscoveryResponse, error) {
	return nil, nil
}
