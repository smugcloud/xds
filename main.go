package main

import (
	"flag"

	api "github.com/smugcloud/xds-consul/consul"
)

func main() {

	flag.Parse()
	api.StartConsul(api.Consul)
	// xds.StartXDS()

}

func init() {
	flag.StringVar(&api.Consul, "Consul address", "http://localhost:8500", "Address of the Consul server.")
	flag.StringVar(&api.Iface, "Address the XDS server will listen on", "127.0.0.1", "Address the XDS server will listen on")
	flag.StringVar(&api.Mgmtport, "Port the XDS server listens on", "19000", "Port the XDS server listens on")
}
