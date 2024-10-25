package main

import (
  "log"
  "net/http"
  "net/url"
  "net/http/httputil"
  "sync"
)

//a backend server
type Server struct{
  URL *url.URL 
  Alive bool 
  ReverseProxy *httputil.ReverseProxy 
}

//server pool holding servers
type ServerPool struct{
  servers []*Server
  current int
  mutex sync.RWMutex 
}

//addserver to the pool
func (sp *ServerPool) AddServer(serverUrl string) error{
  url, err := url.Parse(serverUrl)

  if err != nil{
    return err 
  }

  server := Server{
    URL: url,
    Alive: true,
    ReverseProxy: httputil.NewSingleHostReverseProxy(url),
  }

  sp.mutex.Lock()
  sp.servers = append(sp.servers, &server)
  sp.mutex.Unlock()

  return nil
}

//get next server using round-robin fashion
func(sp *ServerPool) GetNextServer() *Server{
  sp.mutex.Lock()
  defer sp.mutex.Unlock()
  
  //reset counter if we've reached the end
  if sp.current >= len(sp.servers){
    sp.current = 0
  }

  server := sp.servers[sp.current]
  sp.current++

  return server
}

//handle the incoming requests
func (sp *ServerPool) LoadBalanceHandler(w http.ResponseWriter, r *http.Request){
  server := sp.GetNextServer()
  if server != nil{
    server.ReverseProxy.ServeHTTP(w, r)
    return 
  }

  http.Error(w, "No servers availables", http.StatusServiceUnavailable)
}

func main (){
  //create a new server
  serverPool := &ServerPool{
    servers: make([]*Server, 0),
  }

  // Add backend servers
	backends := []string{
		"http://localhost:8081",
		"http://localhost:8082",
		"http://localhost:8083",
	}

  for _, backend := range backends{
    err := serverPool.AddServer(backend)
    if err != nil{
      log.Printf("Error Adding a server %v\n", err)
    }
  }

  //create the load balancer LoadBalanceHandler
  http.HandleFunc("/", serverPool.LoadBalanceHandler)

  //start load balancer at port 8080 
  log.Printf("Starting load balancer on port 8080...\n")
  if err := http.ListenAndServe(":8080", nil); err != nil{
    log.Fatal(err) 
  }
}
