package xds

import (
	"sync"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/gogo/protobuf/proto"
)

// RouteCache manages the contents of the gRPC RDS cache.
type RouteCache struct {
	routeCache
}

type routeCache struct {
	mu      sync.Mutex
	values  map[string]*v2.RouteConfiguration
	waiters []chan int
	last    int
}

func (c *routeCache) Register(ch chan int, last int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if last < c.last {
		// notify this channel immediately
		ch <- c.last
		return
	}
	c.waiters = append(c.waiters, ch)
}

// Update replaces the contents of the cache with the supplied map.
func (c *routeCache) Update(v map[string]*v2.RouteConfiguration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.values = v
	c.notify()
}

// notify notifies all registered waiters that an event has occurred.
func (c *routeCache) notify() {
	c.last++

	for _, ch := range c.waiters {
		ch <- c.last
	}
	c.waiters = c.waiters[:0]
}

// Values returns a slice of the value stored in the cache.
func (c *routeCache) Values(filter func(string) bool) []proto.Message {
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

type routeVisitor struct {
	*RouteCache
}
