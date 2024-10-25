package main

import (
  "log"
  "net/http"
  "net/url"
  "net/http/httputil"
  "sync"
  "time"
  "fmt"
)

//a backend server
type Server struct{
  URL *url.URL 
  Alive bool 
  ReverseProxy *httputil.ReverseProxy 
  mux sync.RWMutex
  failures int
}

func (s *Server) SetAlive(alive bool){
  s.mux.Lock()
  s.Alive = alive
  if alive{
    s.failures = 0
  }

  s.mux.Unlock()
}

func (s *Server) IsAlive()bool{
  s.mux.RLock()
  alive := s.Alive
  s.mux.RUnlock()

  return alive 
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

  //start health checking
  go sp.healthCheck(&server)

  return nil
}



//get next server using round-robin fashion
func(sp *ServerPool) GetNextServer() *Server{
  sp.mutex.Lock()
  defer sp.mutex.Unlock()
  
  //loop until got an alive server of end of servers 
  for i := 0; i < len(sp.servers); i++{
    sp.current = (sp.current + 1) % len(sp.servers)
    server := sp.servers[sp.current]

    if server.IsAlive(){
      return server 
    }
  }
  
  return nil
}

func (sp *ServerPool) healthCheck(server *Server) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	for {
		resp, err := client.Get(fmt.Sprintf("%s/health", server.URL.String()))
		status := false

		if err != nil {
			log.Printf("Health check failed for %s: %v", server.URL, err)
		} else {
			if resp.StatusCode == http.StatusOK {
				status = true
			}
			resp.Body.Close()
		}

		// Update server status
		server.mux.Lock()
		if !status {
			server.failures++
			if server.failures >= 3 { // Mark as dead after 3 consecutive failures
				if server.Alive {
					log.Printf("Server %s is now offline", server.URL)
				}
				server.Alive = false
			}
		} else {
			if !server.Alive {
				log.Printf("Server %s is back online", server.URL)
			}
			server.Alive = true
			server.failures = 0
		}
		server.mux.Unlock()

		// Sleep before next health check
		time.Sleep(10 * time.Second)
	}
}

//handle the incoming requests
func (sp *ServerPool) LoadBalanceHandler(w http.ResponseWriter, r *http.Request){
  //skip health chekcs from being proxied
  if r.URL.Path == "/health" {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Load balancer is healthy\n"))
    return
  }

  server := sp.GetNextServer()
  if server != nil{
    server.ReverseProxy.ServeHTTP(w, r)
    return 
  }

  http.Error(w, "No servers availables", http.StatusServiceUnavailable)
}

//main function
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


