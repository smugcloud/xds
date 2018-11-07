package xds

import (
	"log"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
)

var (
	Consul, Iface, Mgmtport string
)

//Service defines the Service objects we are concerned about
type Service struct {
	Name    string
	Address string
	Port    int
	Dc      string
}

//ConsulCache is the local cache of Services
type ConsulCache struct {
	Services map[string]*Service
	lock     sync.RWMutex
}

//WatchConsul waits for changes to the Consul Services, and updates the cache
func (c *ConsulCache) WatchConsul(client *api.Catalog, qm *api.QueryMeta, qo *api.QueryOptions) {

	qo.WaitIndex = qm.LastIndex
	for {
		log.Print("Waiting for new Services.")
		log.Printf("grpcServer: %+v", c)
		svcs, qm, err := client.Services(qo)
		if err != nil {
			log.Print(err)
		}
		for services := range svcs {
			if _, ok := c.Services[services]; !ok {
				log.Printf("Service does not exist, adding %+v", services)
				// c.Services[services] = 1
				c.lock.Lock()
				log.Printf("Before getting details: %v", c.Services[services])
				c.lock.Unlock()

				sd, nqm, _ := getNewServiceDetails(services, client, qo)
				c.Services[services] = &Service{Name: services, Address: sd.Address, Port: sd.ServicePort}
				qm.LastIndex = nqm.LastIndex

			}
		}
		cs.OnChange(c.Services)

		qo.WaitIndex = qm.LastIndex

	}
}

//StartConsul contacts Consul initially, and builds the grpcServer of objects we are concerned about
func StartConsul(consul string) {
	config := &api.Config{
		Address: consul,
	}
	client, err := api.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	cat := client.Catalog()

	qo := &api.QueryOptions{
		WaitTime: time.Hour,
	}

	svcs, qm, err := cat.Services(qo)
INFINITE:
	for {
		if err != nil {
			log.Printf("Error connecting to Consul.  Retrying in 3 seconds. %v ", err)
			time.Sleep(3 * time.Second)
			continue INFINITE
		}
		break
	}
	cache := NewConsulCache()
	cache.populateCache(svcs)
	cache.getAllServiceDetails(cat, qo)

	go cache.WatchConsul(cat, qm, qo)

}

//Populates the details of new Services which are discovered
func getNewServiceDetails(service string, client *api.Catalog, qo *api.QueryOptions) (*api.CatalogService, *api.QueryMeta, error) {
	svc, qm, err := client.Service(service, "", qo)
	if err != nil {
		log.Print(err)
		return nil, nil, err
	}
	sd := svc[0]
	return sd, qm, nil
}

//Populates the details of each Consul service that we know about in the cache
func (c *ConsulCache) getAllServiceDetails(client *api.Catalog, qo *api.QueryOptions) {
	for n := range c.Services {
		svc, _, _ := client.Service(n, "", qo)
		for _, v := range svc {
			c.Services[n] = &Service{Name: n, Address: v.Address, Port: v.ServicePort}
		}
	}
}

// NewConsulCache creates a new in-memory store
func NewConsulCache() *ConsulCache {
	c := ConsulCache{
		Services: make(map[string]*Service),
	}
	return &c
}

//Populates the initial cache at startup
func (c *ConsulCache) populateCache(svc map[string][]string) {
	for s := range svc {
		c.Services[s] = &Service{Name: s}

	}
}
