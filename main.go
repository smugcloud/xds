package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
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
	httPort = ":19001"
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

func newCoreAddress(port int) *core.Address {
	var coreAddress = &core.Address{Address: &core.Address_SocketAddress{
		SocketAddress: &core.SocketAddress{
			Address:  targetHost,
			Protocol: core.TCP,
			PortSpecifier: &core.SocketAddress_PortValue{
				PortValue: uint32(port),
			},
		},
	}}

	return coreAddress
}

var coreAddresses []*core.Address
var ch chan int

func appendCore(ca *core.Address) *v2.Cluster {
	cluster.Hosts = append(coreAddresses, ca)
	return cluster
}

var cluster = &v2.Cluster{

	Name:            clusterName,
	ConnectTimeout:  2 * time.Second,
	Type:            v2.Cluster_LOGICAL_DNS,
	DnsLookupFamily: v2.Cluster_V4_ONLY,
	LbPolicy:        v2.Cluster_ROUND_ROBIN,
	// Before abstraction
	// Hosts:           []*core.Address{h},
	Hosts: append(coreAddresses, h),
}
var r = route.Route{
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
}

// var v = route.VirtualHost{
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

var v = route.VirtualHost{
	Name:    virtualHostName,
	Domains: []string{"*"},

	Routes: append(routeSlice, r),
}

func newVirtualHost(ro route.Route) route.VirtualHost {
	vh := route.VirtualHost{
		Name:    virtualHostName,
		Domains: []string{"*"},
		Routes:  append(routeSlice, ro),
	}
	return vh
}

var routeSlice []route.Route
var virtualHosts []route.VirtualHost
var manager = &hcm.HttpConnectionManager{
	CodecType:  hcm.AUTO,
	StatPrefix: "ingress_http",
	RouteSpecifier: &hcm.HttpConnectionManager_RouteConfig{
		RouteConfig: &v2.RouteConfiguration{
			Name: "test-route",
			// VirtualHosts: []route.VirtualHost{v},
			VirtualHosts: append(virtualHosts, v),
		},
	},
	HttpFilters: []*hcm.HttpFilter{{
		Name: util.Router,
	}},
}

func appendVhost(rvh route.VirtualHost) {
	manager = &hcm.HttpConnectionManager{
		CodecType:  hcm.AUTO,
		StatPrefix: "ingress_http",
		RouteSpecifier: &hcm.HttpConnectionManager_RouteConfig{
			RouteConfig: &v2.RouteConfiguration{
				Name: "test-route",
				// VirtualHosts: []route.VirtualHost{v},
				VirtualHosts: append(virtualHosts, rvh),
			},
		},
		HttpFilters: []*hcm.HttpFilter{{
			Name: util.Router,
		}},
	}
}

//Path holds the new path we want to add to the cache
type Path struct {
	Path string `json:"path"`
	Port string `json:"port"`
}

//Register is helper to fire the channel
func Register(c chan int) {
	c <- 1
}

func newRoute(prefix string) route.Route {
	route := route.Route{
		Match: route.RouteMatch{
			PathSpecifier: &route.RouteMatch_Prefix{
				Prefix: prefix,
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
	}
	return route
}

//Add a path to the listener server
func addPath(w http.ResponseWriter, req *http.Request) {
	var newPath Path
	body, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(body, &newPath)
	log.Printf("Received: %v", newPath)
	nca := newCoreAddress(9001)
	_ = appendCore(nca)
	log.Printf("Cluster details: %v", cluster)
	// log.Printf("Appended cluster: %v", clus)
	go Register(ch)
	ro := newRoute("/demo")
	vh := newVirtualHost(ro)
	appendVhost(vh)
	log.Printf("virtualHosts slice: %v", virtualHosts)
	lr = &v2.Listener{
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
	dr = &v2.DiscoveryResponse{
		VersionInfo: "number-2",
		Resources: []types.Any{{
			TypeUrl: "type.googleapis.com/envoy.api.v2.Listener",
			Value:   Marshal(lr),
		}},
		TypeUrl: "type.googleapis.com/envoy.api.v2.Listener",
	}
	log.Printf("manager vhost: %v", manager)
	log.Printf("Listener lr: %v", lr)
	go Register(lch)

	//cdr := createClusterDiscoveryResponse(clus)
	// grpc.ServerStream.SendMsg(cdr)
	// v2.ClusterDiscoveryService_StreamClustersServer.Send(cdr)
	log.Printf("Added path: http://localhost:8888/%v", newPath.Path)
}

//Run an HTTP server to demonstrate dynamically updating the MgmtServer configs
func runHTTP() {
	http.HandleFunc("/addPath", addPath)
	log.Print("HTTP Server listening on localhost" + httPort)
	log.Fatal(http.ListenAndServe(httPort, nil))
}

//MessageToStruct takes a PB struct and converts it to the type required by the FilterChain Config
func MessageToStruct(m *hcm.HttpConnectionManager) *types.Struct {
	pbst, err := util.MessageToStruct(m)
	if err != nil {
		panic(err)
	}
	return pbst
}

//Build a Listener Response
var dr = &v2.DiscoveryResponse{
	VersionInfo: "number-2",
	Resources: []types.Any{{
		TypeUrl: "type.googleapis.com/envoy.api.v2.Listener",
		Value:   Marshal(lr),
	}},
	TypeUrl: "type.googleapis.com/envoy.api.v2.Listener",
}

// func createListenerDiscoveryResponse(ldr *v2.Listener) *v2.DiscoveryResponse {
// 	dr = &v2.DiscoveryResponse{
// 		VersionInfo: "number-2",
// 		Resources: []types.Any{{
// 			TypeUrl: "type.googleapis.com/envoy.api.v2.Listener",
// 			Value:   Marshal(ldr),
// 		}},
// 		TypeUrl: "type.googleapis.com/envoy.api.v2.Listener",
// 	}
// 	return dr
// }

//Build a Cluster Response
var cdr = &v2.DiscoveryResponse{
	VersionInfo: "number-2",
	Resources: []types.Any{{
		TypeUrl: "type.googleapis.com/envoy.api.v2.Cluster",
		Value:   MarshalCluster(cluster),
	}},
	TypeUrl: "type.googleapis.com/envoy.api.v2.Cluster",
}

// func createClusterDiscoveryResponse(dr *v2.Cluster) *v2.DiscoveryResponse {
// 	cdr = &v2.DiscoveryResponse{
// 		VersionInfo: "number-2",
// 		Resources: []types.Any{{
// 			TypeUrl: "type.googleapis.com/envoy.api.v2.Cluster",
// 			Value:   MarshalCluster(dr),
// 		}},
// 		TypeUrl: "type.googleapis.com/envoy.api.v2.Cluster",
// 	}
// 	return cdr
// }

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

type clusterServer struct {
}

func main() {

	lis, err := net.Listen(proto, iface+":"+mgrPort)
	if err != nil {
		log.Fatalf("Error getting listener: %v", err)
	}
	log.Printf("Original manager: %v", manager)
	log.Printf("Original virtualHosts: %v", virtualHosts)
	// var srv v2.ListenerDiscoveryServiceServer
	grpc := grpc.NewServer()
	v2.RegisterClusterDiscoveryServiceServer(grpc, &clusterServer{})
	v2.RegisterListenerDiscoveryServiceServer(grpc, &listenerServer{})

	go runHTTP()
	log.Printf("gRPC Listening on: %v", "http://"+iface+":"+mgrPort)
	grpc.Serve(lis)
}

var last = 0

func (c *clusterServer) StreamClusters(scs v2.ClusterDiscoveryService_StreamClustersServer) error {
	log.Print("In Cluster Stream")
	ch = make(chan int, 1)
	//Fill the channel to send initial cluster
	ch <- 1
	for {
		select {
		case last = <-ch:
			log.Print("Channel fired.\n")
			ds, err := scs.Recv()
			log.Printf("Recieved: %v\n", ds)
			log.Printf("Sending DiscoveryResponse: %v\n", cdr)
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

var lch chan int

func (l *listenerServer) StreamListeners(sls v2.ListenerDiscoveryService_StreamListenersServer) error {

	log.Print("In the stream")
	lch = make(chan int, 1)
	//Fill the channel to send initial cluster
	lch <- 1
	for {
		select {
		case last = <-lch:
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
