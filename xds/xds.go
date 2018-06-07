package xds

import (
	"log"
	"time"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	"github.com/envoyproxy/go-control-plane/pkg/util"

	prot "github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
)

//UberRoutes is the simple struct to store all of our things (for now)
type UberRoutes struct {
	Clusters      []*v2.Cluster
	CoreAddresses []*core.Address
	VHost         []route.VirtualHost
	Routes        []route.Route
}

func NewRoute(path, cluster string) route.Route {
	route := route.Route{
		Match: route.RouteMatch{
			PathSpecifier: &route.RouteMatch_Prefix{
				Prefix: path,
			},
		},
		Action: &route.Route_Route{
			Route: &route.RouteAction{
				// HostRewriteSpecifier: &route.RouteAction_HostRewrite{
				// 	HostRewrite: targetHost,
				// },
				ClusterSpecifier: &route.RouteAction_Cluster{
					Cluster: cluster,
				},
			},
		},
	}
	return route
}

// func (u *UberRoutes) Playground() {
// 	log.Printf("UberRoutes in Playground: %v", u)
// }

//NewCluster creates a Cluster object
func (u *UberRoutes) NewCluster(address, name string, port int) []*v2.Cluster {
	// log.Printf("Length of CoreAddresses: %v", len(u.CoreAddresses))
	var cluster = &v2.Cluster{

		Name:            name,
		ConnectTimeout:  2 * time.Second,
		Type:            v2.Cluster_LOGICAL_DNS,
		DnsLookupFamily: v2.Cluster_V4_ONLY,
		LbPolicy:        v2.Cluster_ROUND_ROBIN,
		// Hosts:           u.CoreAddresses,
		Hosts: []*core.Address{{Address: &core.Address_SocketAddress{
			SocketAddress: &core.SocketAddress{
				Address:  address,
				Protocol: core.TCP,
				PortSpecifier: &core.SocketAddress_PortValue{
					PortValue: uint32(port),
				},
			},
		},
		}},
	}

	u.Clusters = append(u.Clusters, cluster)
	return u.Clusters
}

//NewCoreAddress creates a core.Address
func (u *UberRoutes) NewCoreAddress(host string, port int) *core.Address {
	var coreAddress = &core.Address{Address: &core.Address_SocketAddress{
		SocketAddress: &core.SocketAddress{
			Address:  host,
			Protocol: core.TCP,
			PortSpecifier: &core.SocketAddress_PortValue{
				PortValue: uint32(port),
			},
		},
	}}

	return coreAddress
}

func (u *UberRoutes) AppendCoreAddress(c *core.Address) []*core.Address {
	u.CoreAddresses = append(u.CoreAddresses, c)
	log.Printf("After appending to CoreAddress: %v", u.CoreAddresses)
	return u.CoreAddresses
}

func (u *UberRoutes) UpdateVirtualHost(vhn string) {
	vh := route.VirtualHost{
		Name:    vhn,
		Domains: []string{"*"},
		Routes:  u.Routes,
	}
	log.Printf("vh in UpdateVirtualHost: %s", vh)
	u.VHost = nil
	u.VHost = append(u.VHost, vh)

}

//NewVirtualHost creates a VirtualHost
func (u *UberRoutes) NewVirtualHost(vhn string) route.VirtualHost {
	return route.VirtualHost{
		Name:    vhn,
		Domains: []string{"*"},
		Routes:  u.Routes,
	}

}

func (u *UberRoutes) AppendVhost(vh route.VirtualHost) {
	u.VHost = append(u.VHost, vh)
}

func (u *UberRoutes) AppendRoute(rt route.Route) {
	u.Routes = append(u.Routes, rt)
}

func (u *UberRoutes) NewHCMAndListener() *v2.Listener {
	manager := &hcm.HttpConnectionManager{
		CodecType:  hcm.AUTO,
		StatPrefix: "ingress_http",
		RouteSpecifier: &hcm.HttpConnectionManager_RouteConfig{
			RouteConfig: &v2.RouteConfiguration{
				Name:         "test-route",
				VirtualHosts: u.VHost,
			},
		},
		HttpFilters: []*hcm.HttpFilter{{
			Name: util.Router,
		}},
	}
	lr := &v2.Listener{
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
	return lr
}

//BootstrapListener creates the initial Listener to load the first route
func (u *UberRoutes) BootstrapListener(clusterName string) *v2.Listener {
	// var uber UberRoutes
	nr := NewRoute("/probe", clusterName)
	u.AppendRoute(nr)
	log.Printf("Original routes: %s", u.Routes)
	vh := u.NewVirtualHost("demo")
	u.AppendVhost(vh)
	lr := u.NewHCMAndListener()
	return lr
}

//MessageToStruct takes a PB struct and converts it to the type required by the FilterChain Config
func MessageToStruct(m *hcm.HttpConnectionManager) *types.Struct {
	pbst, err := util.MessageToStruct(m)
	if err != nil {
		panic(err)
	}
	return pbst
}

// toAny converts the contens of a resourcer's Values to the
// respective slice of types.Any.
func toAny(dr []*v2.Cluster) ([]types.Any, error) {
	log.Printf("Length of slice: %v", len(dr))
	typeUrl := "type.googleapis.com/envoy.api.v2.Cluster"
	resources := make([]types.Any, len(dr))
	for i := range dr {
		value := MarshalCluster(dr[i])
		resources[i] = types.Any{TypeUrl: typeUrl, Value: value}
	}
	log.Printf("Final resources before sent: %s", resources)
	return resources, nil
}

func CreateClusterDiscoveryResponse(dr []*v2.Cluster) *v2.DiscoveryResponse {
	gettoAny, err := toAny(dr)
	if err != nil {
		log.Fatal(err)
	}

	cdr := &v2.DiscoveryResponse{
		VersionInfo: "number-2",
		Resources:   gettoAny,
		TypeUrl:     "type.googleapis.com/envoy.api.v2.Cluster",
	}
	return cdr
}

func CreateListenerDiscoveryResponse(ldr *v2.Listener) *v2.DiscoveryResponse {
	dr := &v2.DiscoveryResponse{
		VersionInfo: "number-2",
		Resources: []types.Any{{
			TypeUrl: "type.googleapis.com/envoy.api.v2.Listener",
			Value:   Marshal(ldr),
		}},
		TypeUrl: "type.googleapis.com/envoy.api.v2.Listener",
	}
	return dr
}

//MarshalCluster marshal's a cluster into a byte slice
func MarshalCluster(l *v2.Cluster) []byte {

	m, err := prot.Marshal(l)
	if err != nil {
		log.Fatal(err)
	}
	return m

}

//Marshal creates a byte slice from a Message
func Marshal(l *v2.Listener) []byte {

	m, _ := prot.Marshal(l)
	return m

}
