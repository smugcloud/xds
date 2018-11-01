package consul

import (
	"log"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
)

var (
	Consul, Iface, Mgmtport string
)

type Service struct {
	Name    string
	Address string
	Port    int
	Dc      string
}

type Cache struct {
	// // Services       map[string]int
	// ServiceDetails []Service
	Services map[string]*Service
	lock     sync.RWMutex
}

func (c *Cache) WatchConsul(client *api.Catalog, qm *api.QueryMeta, qo *api.QueryOptions) {
	qo.WaitIndex = qm.LastIndex
	for {
		// svcs, qm, err := cat.ServiceDetails(qo)
		// if err != nil {
		// 	log.Printf("Error connecting to Consul.  Retrying in 3 seconds. %v ", err)
		// 	time.Sleep(3 * time.Second)
		// 	continue
		// }
		// log.Printf("%+v ; %+v", svcs, qm)
		log.Print("Waiting for new Services.")
		log.Printf("Cache: %+v", c)
		// qo.WaitIndex = qm.LastIndex
		svcs, qm, err := client.Services(qo)
		if err != nil {
			log.Print(err)
		}
		// for services := range c.Services {
		// 	log.Printf("Comparing %+v and %+v", services, svcs)
		// 	if _, ok := svcs[services]; !ok {
		// 		log.Printf("Service does not exist, adding %+v", services)
		// 	}
		// }
		for services := range svcs {
			if _, ok := c.Services[services]; !ok {
				log.Printf("Service does not exist, adding %+v", services)
				// c.Services[services] = 1
				c.lock.Lock()
				c.Services[services] = &Service{Name: services}
				c.lock.Unlock()

				sd, nqm, _ := getNewServiceDetails(services, client, qo)
				c.Services[services] = &Service{Address: sd.Address, Port: sd.ServicePort}
				qm.LastIndex = nqm.LastIndex
				log.Printf("Updated cache: %+v", c.Services)
			}
		}

		qo.WaitIndex = qm.LastIndex

	}
}

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
	// fmt.Printf("ServiceDetails: %+v\n", len(svcs))
	if err != nil {
		log.Print(err)
	}
	cache := NewCache()
	cache.populateCache(svcs)
	cache.getAllServiceDetails(cat, qo)

	cache.WatchConsul(cat, qm, qo)
	// fmt.Printf("%+v\n", cache)
	// svc := BuildService()
	// RegisterService(svc)
}

func getNewServiceDetails(service string, client *api.Catalog, qo *api.QueryOptions) (*api.CatalogService, *api.QueryMeta, error) {
	svc, qm, err := client.Service(service, "", qo)
	if err != nil {
		log.Print(err)
		return nil, nil, err
	}
	sd := svc[0]
	return sd, qm, nil
}

func (c *Cache) getAllServiceDetails(client *api.Catalog, qo *api.QueryOptions) *Cache {
	// for k, n := range c.ServiceDetails {
	// 	svc, _, _ := client.Service(n.Name, "", qo)
	// 	for _, v := range svc {
	// 		c.lock.Lock()
	// 		c.ServiceDetails[k].Address = v.ServiceAddress
	// 		c.ServiceDetails[k].Port = v.ServicePort
	// 		c.lock.Unlock()
	// 		// fmt.Printf("%+v\n", v)
	// 	}

	// }

	for n := range c.Services {
		svc, _, _ := client.Service(n, "", qo)
		for _, v := range svc {
			c.Services[n] = &Service{Address: v.Address, Port: v.ServicePort}
		}

	}
	return c
}

func NewCache() *Cache {
	c := Cache{
		// Services:       make(map[string]int),
		// ServiceDetails: make([]Service, 0),
		Services: make(map[string]*Service),
	}
	return &c
}

func (c *Cache) populateCache(svc map[string][]string) *Cache {
	for s := range svc {
		// c.Services[s] = 1
		// c.ServiceDetails = append(c.ServiceDetails, Service{Name: s})
		c.Services[s] = &Service{Name: s}

	}
	log.Printf("Cache after initial population: %+v", c)
	return c
}

//BuildService just creates a static service we can use
func BuildService() api.AgentServiceRegistration {
	return api.AgentServiceRegistration{
		ID:      "probe",
		Name:    "probe",
		Address: "127.0.0.1",
		Port:    9000,
	}
}

func RegisterService(svc api.AgentServiceRegistration) {
	// Get a new client
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		panic(err)
	}
	agent := client.Agent()
	agent.ServiceRegister(&svc)

}
