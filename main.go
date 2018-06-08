package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	"github.com/smugcloud/xds/xds"
	"google.golang.org/grpc"
)

const (
	proto   = "tcp"
	iface   = "127.0.0.1"
	mgrPort = "19000"
	httPort = ":19001"
)

var listenerName = "nick-xds"
var targetPrefix = "/probe"
var virtualHostName = "local_service"
var targetHost = "127.0.0.1"
var clusterName = "nick-xds"
var secondCluster = "port9001"

var vhName = "demo"

type listenerServer struct {
}

// var coreAddresses []*core.Address
var ch chan int
var routeSlice []route.Route
var virtualHosts []route.VirtualHost

//Path holds the new path we want to add to the cache
type Path struct {
	Path string `json:"path"`
	Port string `json:"port"`
}

//Register is helper to fire the channel
func Register(c chan int) {
	c <- 1
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

	newUber.Clusters = newUber.NewCluster(targetHost, secondCluster, 9001)

	//Build the new route.Route
	nr := xds.NewRoute(newPath.Path, secondCluster)

	//Append the new route to the slice
	newUber.AppendRoute(nr)

	//Updates the routes in VirtualHost
	newUber.UpdateVirtualHost(vhName)

	bl := newUber.NewHCMAndListener()

	cdr = xds.CreateClusterDiscoveryResponse(newUber.Clusters)

	Register(ch)
	ldr = xds.CreateListenerDiscoveryResponse(bl)

	go Register(lch)
	log.Printf("Added path: http://127.0.0.1:8888/%v", newPath.Path)
}

//Run an HTTP server to demonstrate dynamically updating the MgmtServer configs
func runHTTP() {
	http.HandleFunc("/addPath", addPath)
	log.Print("HTTP Server listening on http://127.0.0.1" + httPort)
	log.Fatal(http.ListenAndServe(httPort, nil))
}

type clusterServer struct {
}

var (
	newUber  = xds.UberRoutes{}
	cdr, ldr *v2.DiscoveryResponse
)

func main() {

	newUber.Clusters = newUber.NewCluster(targetHost, clusterName, 9000)

	//Build the initial CDR
	cdr = xds.CreateClusterDiscoveryResponse(newUber.Clusters)

	bl := newUber.BootstrapListener(clusterName)
	//Build the initial LDR
	ldr = xds.CreateListenerDiscoveryResponse(bl)

	lis, err := net.Listen(proto, iface+":"+mgrPort)
	if err != nil {
		log.Fatalf("Error getting listener: %v", err)
	}

	grpc := grpc.NewServer()
	v2.RegisterClusterDiscoveryServiceServer(grpc, &clusterServer{})
	v2.RegisterListenerDiscoveryServiceServer(grpc, &listenerServer{})

	go runHTTP()
	log.Printf("gRPC Listening on: %v", "http://"+iface+":"+mgrPort)
	grpc.Serve(lis)
}

var last = 0

func (c *clusterServer) StreamClusters(scs v2.ClusterDiscoveryService_StreamClustersServer) error {
	ch = make(chan int, 1)
	//Fill the channel to send initial cluster
	ch <- 1
	for {
		select {
		case last = <-ch:
			// log.Print("Channel fired.\n")
			_, err := scs.Recv()
			// log.Printf("Recieved: %v\n", ds)
			// log.Printf("Sending DiscoveryResponse: %v\n", cdr)
			log.Print("Sending CDS Response.")
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

	// log.Print("In the stream")
	lch = make(chan int, 1)
	//Fill the channel to send initial cluster
	lch <- 1
	for {
		select {
		case last = <-lch:
			_, err := sls.Recv()
			// log.Printf("Recieved: %v\n", ds)
			// log.Printf("Sending DiscoveryResponse: %v\n", ldr)
			log.Print("Sending LDS Response.")
			sls.Send(ldr)

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
