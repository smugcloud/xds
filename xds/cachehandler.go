package xds

import (
	"log"
	"time"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
)

// CacheHandler manages the state of xDS caches.
type CacheHandler struct {
	ListenerCache
	RouteCache
	ClusterCache
}

//AddCluster appends to our cluster object
func (ch *CacheHandler) UpdateEnvoy(name string) {

	var cluster = &v2.Cluster{
		Name:            name,
		ConnectTimeout:  2 * time.Second,
		Type:            v2.Cluster_LOGICAL_DNS,
		DnsLookupFamily: v2.Cluster_V4_ONLY,
		LbPolicy:        v2.Cluster_ROUND_ROBIN,
		Hosts: []*core.Address{{Address: &core.Address_SocketAddress{
			SocketAddress: &core.SocketAddress{
				Address:  "127.0.0.1",
				Protocol: core.TCP,
				PortSpecifier: &core.SocketAddress_PortValue{
					PortValue: uint32(9000),
				},
			},
		},
		}},
	}
	clusters := map[string]*v2.Cluster{
		name: cluster,
	}
	ch.ClusterCache.Update(clusters)
}

func (ch *CacheHandler) OnChange(s map[string]*Service) {
	clusters := map[string]*v2.Cluster{}

	for k, v := range s {
		log.Printf("key: %v, value: %v", k, v)
		cluster := &v2.Cluster{
			Name:            v.Name,
			ConnectTimeout:  2 * time.Second,
			Type:            v2.Cluster_LOGICAL_DNS,
			DnsLookupFamily: v2.Cluster_V4_ONLY,
			LbPolicy:        v2.Cluster_ROUND_ROBIN,
			Hosts: []*core.Address{{Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Address:  v.Address,
					Protocol: core.TCP,
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: uint32(v.Port),
					},
				},
			},
			}},
		}
		clusters[k] = cluster
	}
	ch.clusterCache.Update(clusters)
}
