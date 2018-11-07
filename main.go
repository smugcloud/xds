package main

import (
	"flag"

	"github.com/smugcloud/xds-consul/bootstrap"
	"github.com/smugcloud/xds-consul/xds"
)

func main() {

	flag.Parse()
	s := xds.InitializeCache()
	xds.StartConsul(bootstrap.Consul)
	xds.StartXDS(s)

}

func init() {
	flag.StringVar(&bootstrap.Consul, "Consul address", "http://localhost:8500", "Address of the Consul server.")
	flag.StringVar(&bootstrap.Iface, "Address the XDS server will listen on", "127.0.0.1", "Address the XDS server will listen on")
	flag.StringVar(&bootstrap.Mgmtport, "Port the XDS server listens on", "19000", "Port the XDS server listens on")
}
