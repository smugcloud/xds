package main

import (
	"context"
	"io"
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
			PortValue: uint32(443),
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

//Marshal creates a byte slice from a Message
func Marshal(l *v2.Listener) []byte {

	m, _ := prot.Marshal(l)
	return m

}

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

var cdr = &v2.DiscoveryResponse{
	VersionInfo: "number-2",
	Resources: []types.Any{{
		TypeUrl: "type.googleapis.com/envoy.api.v2.Cluster",
		Value:   MarshalCluster(cluster),
	}},
	TypeUrl: "type.googleapis.com/envoy.api.v2.Cluster",
}

func (c *clusterServer) StreamClusters(scs v2.ClusterDiscoveryService_StreamClustersServer) error {
	log.Print("In Cluster Stream")
	ds, err := scs.Recv()
	log.Printf("Recieved: %v\n", ds)
	log.Printf("Sending DiscoveryResponse: %v\n", dr)
	scs.Send(cdr)
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	return nil
}

func (c *clusterServer) FetchClusters(context.Context, *v2.DiscoveryRequest) (*v2.DiscoveryResponse, error) {
	return nil, nil
}

func (l *listenerServer) StreamListeners(sls v2.ListenerDiscoveryService_StreamListenersServer) error {

	log.Print("In the stream")
	// for {
	ds, err := sls.Recv()
	log.Printf("Recieved: %v\n", ds)
	log.Printf("Sending DiscoveryResponse: %v\n", dr)
	sls.Send(dr)
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	//}
	return nil
}

func (l *listenerServer) FetchListeners(context.Context, *v2.DiscoveryRequest) (*v2.DiscoveryResponse, error) {
	return nil, nil
}

// create a cache
// signal := make(chan struct{})
// cb := &callbacks{signal: signal}
// config := cache.NewSnapshotCache(false, test.Hasher{}, logger{})
// srv := server.NewServer(config, cb)
// grpc := grpc.NewServer()
// v2.RegisterListenerDiscoveryServiceServer(grpc, srv)
// v2.RegisterClusterDiscoveryServiceServer(grpc, srv)
// lis, err := net.Listen("tcp", "localhost:19000")
// if err != nil {
// 	log.Printf("Error acquiring listner: %v", err)
// }
// atomic.AddInt32(&version, 1)

// var listenerName = "nick-xds"
// var targetHost = "www.bbc.com"
// var targetPrefix = "/robots.txt"
// var virtualHostName = "local_service"
// var remoteHost = "www.bbc.com"

// var routeConfigName = "local_route"
// nodeID := "test-id"
// clusterName := "nick-xds"
// h := &core.Address{Address: &core.Address_SocketAddress{
// 	SocketAddress: &core.SocketAddress{
// 		Address:  remoteHost,
// 		Protocol: core.TCP,
// 		PortSpecifier: &core.SocketAddress_PortValue{
// 			PortValue: uint32(443),
// 		},
// 	},
// }}
// c := []cache.Resource{
// 	&v2.Cluster{
// 		Name:            clusterName,
// 		ConnectTimeout:  2 * time.Second,
// 		Type:            v2.Cluster_LOGICAL_DNS,
// 		DnsLookupFamily: v2.Cluster_V4_ONLY,
// 		LbPolicy:        v2.Cluster_ROUND_ROBIN,
// 		Hosts:           []*core.Address{h},
// 	},
// }

// v := route.VirtualHost{
// 	Name:    virtualHostName,
// 	Domains: []string{"*"},

// 	Routes: []route.Route{{
// 		Match: route.RouteMatch{
// 			PathSpecifier: &route.RouteMatch_Prefix{
// 				Prefix: targetPrefix,
// 			},
// 		},
// 		Action: &route.Route_Route{
// 			Route: &route.RouteAction{
// 				HostRewriteSpecifier: &route.RouteAction_HostRewrite{
// 					HostRewrite: targetHost,
// 				},
// 				ClusterSpecifier: &route.RouteAction_Cluster{
// 					Cluster: clusterName,
// 				},
// 			},
// 		},
// 	}}}
// manager := &v2.HttpConnectionManager{
// 	CodecType:  v2.AUTO,
// 	StatPrefix: "ingress_http",
// 	RouteSpecifier: &v2.HttpConnectionManager_RouteConfig{
// 		RouteConfig: &v2.RouteConfiguration{
// 			Name:         routeConfigName,
// 			VirtualHosts: []route.VirtualHost{v},
// 		},
// 	},
// 	HttpFilters: []*v2.HttpFilter{{
// 		Name: util.Router,
// 	}},
// }
// pbst, err := util.MessageToStruct(manager)
// if err != nil {
// 	panic(err)
// }

// var l = []cache.Resource{
// 	&v2.Listener{
// 		Name: listenerName,
// 		Address: core.Address{
// 			Address: &core.Address_SocketAddress{
// 				SocketAddress: &core.SocketAddress{
// 					Protocol: core.TCP,
// 					Address:  localhost,
// 					PortSpecifier: &core.SocketAddress_PortValue{
// 						PortValue: 80,
// 					},
// 				},
// 			},
// 		},
// 		FilterChains: []listener.FilterChain{{
// 			Filters: []listener.Filter{{
// 				Name:   util.HTTPConnectionManager,
// 				Config: pbst,
// 			}},
// 		}},
// 	}}
// log.Infof(">>>>>>>>>>>>>>>>>>> creating snapshot Version 0")
// snap := cache.NewSnapshot(string(version), nil, c, nil, l)

// config.SetSnapshot(nodeID, snap)
// grpc.Serve(lis)
// }

// type logger struct{}

// func (logger logger) Infof(format string, args ...interface{}) {
// log.Debugf(format, args...)
// }
// func (logger logger) Errorf(format string, args ...interface{}) {
// log.Errorf(format, args...)
// }

// // Hasher returns node ID as an ID
// type Hasher struct {
// }

// type callbacks struct {
// signal   chan struct{}
// fetches  int
// requests int
// mu       sync.Mutex
// }

// func (cb *callbacks) Report() {
// cb.mu.Lock()
// defer cb.mu.Unlock()
// log.WithFields(log.Fields{"fetches": cb.fetches, "requests": cb.requests}).Info("server callbacks")
// }
// func (cb *callbacks) OnStreamOpen(id int64, typ string) {
// log.Debugf("stream %d open for %s", id, typ)
// }
// func (cb *callbacks) OnStreamClosed(id int64) {
// log.Debugf("stream %d closed", id)
// }
// func (cb *callbacks) OnStreamRequest(int64, *v2.DiscoveryRequest) {
// cb.mu.Lock()
// defer cb.mu.Unlock()
// cb.requests++
// if cb.signal != nil {
// 	close(cb.signal)
// 	cb.signal = nil
// }
// }
// func (cb *callbacks) OnStreamResponse(int64, *v2.DiscoveryRequest, *v2.DiscoveryResponse) {}
// func (cb *callbacks) OnFetchRequest(req *v2.DiscoveryRequest) {
// cb.mu.Lock()
// defer cb.mu.Unlock()
// cb.fetches++
// if cb.signal != nil {
// 	close(cb.signal)
// 	cb.signal = nil
// }
// }
// func (cb *callbacks) OnFetchResponse(*v2.DiscoveryRequest, *v2.DiscoveryResponse) {}
