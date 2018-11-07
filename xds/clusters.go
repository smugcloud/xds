package xds

import (
	"log"
	"sync"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/gogo/protobuf/proto"
)

// ClusterCache manages the contents of the gRPC CDS cache.
type ClusterCache struct {
	clusterCache
}

type clusterCache struct {
	mu      sync.Mutex
	values  map[string]*v2.Cluster
	waiters []chan int
	last    int
}

// Register registers ch to receive a value when Notify is called.
func (c *clusterCache) Register(ch chan int, last int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	log.Print("clusterCache registered.")
	if last < c.last {
		// notify this channel immediately
		ch <- c.last
		return
	}
	c.waiters = append(c.waiters, ch)
}

// Update replaces the contents of the cache with the supplied map.
func (c *clusterCache) Update(v map[string]*v2.Cluster) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.values = v
	c.notify()
}

// notify notifies all registered waiters that an event has occurred.
func (c *clusterCache) notify() {
	c.last++

	for _, ch := range c.waiters {
		ch <- c.last
	}
	c.waiters = c.waiters[:0]
}

// Values returns a slice of the value stored in the cache.
func (c *clusterCache) Values(filter func(string) bool) []proto.Message {
	c.mu.Lock()
	values := make([]proto.Message, 0, len(c.values))
	for _, v := range c.values {
		if filter(v.Name) {
			values = append(values, v)
		}
	}
	c.mu.Unlock()
	return values
}

// clusterVisitor walks a *dag.DAG and produces a map of *v2.Clusters.
type clusterVisitor struct {
	*ClusterCache

	clusters map[string]*v2.Cluster
}

func (v *clusterVisitor) Visit() map[string]*v2.Cluster {
	v.clusters = make(map[string]*v2.Cluster)
	// v.Visitable.Visit(v.visit)
	return v.clusters
}
