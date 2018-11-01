package xds

import (
	"log"
	"net"

	api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	xds "github.com/envoyproxy/go-control-plane/pkg/server"
	consul "github.com/smugcloud/xds-consul/consul"
	"google.golang.org/grpc"
)

// Hasher returns node ID as an ID
type Hasher struct {
}

// ID function
func (h Hasher) ID(node *core.Node) string {
	if node == nil {
		return "unknown"
	}
	return node.Id
}

//StartXDS starts the management server for Envoy to connect to.
func StartXDS() {
	lis, err := net.Listen("tcp", consul.Iface+":"+consul.Mgmtport)
	if err != nil {
		log.Print(err)
	}
	snapshotCache := cache.NewSnapshotCache(false, Hasher{}, nil)
	server := xds.NewServer(snapshotCache, nil)
	grpcServer := grpc.NewServer()

	discovery.RegisterAggregatedDiscoveryServiceServer(grpcServer, server)
	api.RegisterEndpointDiscoveryServiceServer(grpcServer, server)
	api.RegisterClusterDiscoveryServiceServer(grpcServer, server)
	api.RegisterRouteDiscoveryServiceServer(grpcServer, server)
	api.RegisterListenerDiscoveryServiceServer(grpcServer, server)

	log.Printf("XDS listening on %v:%v", consul.Iface, consul.Mgmtport)

	if err := grpcServer.Serve(lis); err != nil {
		log.Print(err)
	}

}
