package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type Server interface {
	Address() string
	isAlive() bool
	Serve(w http.ResponseWriter, r *http.Request)
}

type simpleServer struct {
	addr string
	proxy *httputil.ReverseProxy
}

func newSimpleServer(addr string) *simpleServer {
	serverUrl, err := url.Parse(addr)
	handleErr(err)

	return &simpleServer{
		addr: addr,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),

	}
}

type LoadBalancer struct {
	port string
	servers []Server
	roundRobinCount int
}

func NewLoadBalancer(port string, servers []Server) *LoadBalancer {
	return &LoadBalancer{
		port: port,
		roundRobinCount: 0,
		servers: servers,
	}
}

func handleErr(err error) {
	if err != nil {
		fmt.Printf("error: %v", err)
		os.Exit(1)
	}
}

func (s *simpleServer) Address() string {return s.addr}

func (s *simpleServer) isAlive() bool {return true}

func (s *simpleServer) Serve(w http.ResponseWriter, r *http.Request) {
	s.proxy.ServeHTTP(w,r)
} 

func (lb *LoadBalancer) getNextAvailableServer() Server{
	server := lb.servers[lb.roundRobinCount%len(lb.servers)]
	for !server.isAlive() {
		lb.roundRobinCount ++
		server = lb.servers[lb.roundRobinCount%len(lb.servers)]
	}
	lb.roundRobinCount++
	return server
}

func (lb *LoadBalancer) serverProxy(w http.ResponseWriter, r *http.Request){
	targetServer := lb.getNextAvailableServer()
	fmt.Printf("forwarding request to address %q\n", targetServer.Address())
	targetServer.Serve(w,r)
}

func main(){
	servers := []Server{
		newSimpleServer("https://www.hiteshjainn.co"),
		newSimpleServer("https://www.instagram.com"),
		newSimpleServer("http://www.duckduckgo.com"),
	}

	lb := NewLoadBalancer("8000", servers)
	
	handleRedirect := func (w http.ResponseWriter, r *http.Request)  {
		lb.serverProxy(w,r)
	}

	http.HandleFunc("/", handleRedirect)
	fmt.Printf("server running at localhost:%s", lb.port)
	http.ListenAndServe(":"+lb.port, nil)
}