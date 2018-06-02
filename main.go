package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	"github.com/envoyproxy/go-control-plane/pkg/util"
	prot "github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"google.golang.org/grpc"
)

const (
	proto   = "tcp"
	iface   = "localhost"
	mgrPort = "19000"
)

var listenerName = "nick-xds"
var targetPrefix = "/probe"
var virtualHostName = "local_service"
var targetHost = "127.0.0.1"
var clusterName = "nick-xds"

var lr = &v2.Listener{
	Name: "listener-0",
	Address: core.Address{
		Address: &core.Address_SocketAddress{
			SocketAddress: &core.SocketAddress{
				Protocol: core.TCP,
				Address:  "127.0.0.1",
				PortSpecifier: &core.SocketAddress_PortValue{
					PortValue: 8888,
				},
			},
		},
	},
	FilterChains: []listener.FilterChain{{
		Filters: []listener.Filter{{
			Name:   util.HTTPConnectionManager,
			Config: MessageToStruct(manager),
		}},
	}},
}

type listenerServer struct {
}

var h = &core.Address{Address: &core.Address_SocketAddress{
	SocketAddress: &core.SocketAddress{
		Address:  targetHost,
		Protocol: core.TCP,
		PortSpecifier: &core.SocketAddress_PortValue{
			PortValue: uint32(9000),
		},
	},
}}

var cluster = &v2.Cluster{

	Name:            clusterName,
	ConnectTimeout:  2 * time.Second,
	Type:            v2.Cluster_LOGICAL_DNS,
	DnsLookupFamily: v2.Cluster_V4_ONLY,
	LbPolicy:        v2.Cluster_ROUND_ROBIN,
	Hosts:           []*core.Address{h},
}

var v = route.VirtualHost{
	Name:    virtualHostName,
	Domains: []string{"*"},

	Routes: []route.Route{{
		Match: route.RouteMatch{
			PathSpecifier: &route.RouteMatch_Prefix{
				Prefix: targetPrefix,
			},
		},
		Action: &route.Route_Route{
			Route: &route.RouteAction{
				HostRewriteSpecifier: &route.RouteAction_HostRewrite{
					HostRewrite: targetHost,
				},
				ClusterSpecifier: &route.RouteAction_Cluster{
					Cluster: clusterName,
				},
			},
		},
	}}}
var manager = &hcm.HttpConnectionManager{
	CodecType:  hcm.AUTO,
	StatPrefix: "ingress_http",
	RouteSpecifier: &hcm.HttpConnectionManager_RouteConfig{
		RouteConfig: &v2.RouteConfiguration{
			Name:         "test-route",
			VirtualHosts: []route.VirtualHost{v},
		},
	},
	HttpFilters: []*hcm.HttpFilter{{
		Name: util.Router,
	}},
}

//MessageToStruct takes a PB struct and converts it to the type required by the FilterChain Config
func MessageToStruct(m *hcm.HttpConnectionManager) *types.Struct {
	pbst, err := util.MessageToStruct(m)
	if err != nil {
		panic(err)
	}
	return pbst
}

//Push a DiscoveryResponse
var dr = &v2.DiscoveryResponse{
	VersionInfo: "number-2",
	Resources: []types.Any{{
		TypeUrl: "type.googleapis.com/envoy.api.v2.Listener",
		Value:   Marshal(lr),
	}},
	TypeUrl: "type.googleapis.com/envoy.api.v2.Listener",
}

var cdr = &v2.DiscoveryResponse{
	VersionInfo: "number-2",
	Resources: []types.Any{{
		TypeUrl: "type.googleapis.com/envoy.api.v2.Cluster",
		Value:   MarshalCluster(cluster),
	}},
	TypeUrl: "type.googleapis.com/envoy.api.v2.Cluster",
}

//Marshal creates a byte slice from a Message
func Marshal(l *v2.Listener) []byte {

	m, _ := prot.Marshal(l)
	return m

}

//MarshalCluster marshal's a cluster into a byte slice
func MarshalCluster(l *v2.Cluster) []byte {

	m, _ := prot.Marshal(l)
	return m

}

type clusterServer struct{}

func main() {

	lis, err := net.Listen(proto, iface+":"+mgrPort)
	if err != nil {
		log.Fatalf("Error getting listener: %v", err)
	}
	// var srv v2.ListenerDiscoveryServiceServer
	grpc := grpc.NewServer()
	v2.RegisterClusterDiscoveryServiceServer(grpc, &clusterServer{})
	v2.RegisterListenerDiscoveryServiceServer(grpc, &listenerServer{})
	log.Printf("Listening on: %v", "http://"+iface+":"+mgrPort)
	grpc.Serve(lis)
}

var last = 0

func (c *clusterServer) StreamClusters(scs v2.ClusterDiscoveryService_StreamClustersServer) error {
	log.Print("In Cluster Stream")
	ch := make(chan int, 1)
	//Fill the channel to send initial cluster
	ch <- 1
	for {
		select {
		case last = <-ch:
			ds, err := scs.Recv()
			log.Printf("Recieved: %v\n", ds)
			log.Printf("Sending DiscoveryResponse: %v\n", dr)
			scs.Send(cdr)

			if err != nil {
				return err
			}
		default:
		}
	}
}

func (c *clusterServer) FetchClusters(context.Context, *v2.DiscoveryRequest) (*v2.DiscoveryResponse, error) {
	return nil, nil
}

func (l *listenerServer) StreamListeners(sls v2.ListenerDiscoveryService_StreamListenersServer) error {

	log.Print("In the stream")
	ch := make(chan int, 1)
	//Fill the channel to send initial cluster
	ch <- 1
	for {
		select {
		case last = <-ch:
			ds, err := sls.Recv()
			log.Printf("Recieved: %v\n", ds)
			log.Printf("Sending DiscoveryResponse: %v\n", dr)
			sls.Send(dr)

			if err != nil {
				return err
			}
		default:
		}
	}
}

func (l *listenerServer) FetchListeners(context.Context, *v2.DiscoveryRequest) (*v2.DiscoveryResponse, error) {
	return nil, nil
}
