package main

import (
  "fmt"
  "os"
  "net/http"
  "flag"
  "log"
  "sync/atomic"
)

//use atomic operations to handle concurrent access
var ( health int32 = 1 ) //1 healthy //0 unhealthy

func main(){
  //define commandline flags for port numbers
  port := flag.Int("port", 8081, "port to server on")
  flag.Parse()

  //create a uniques server identifier
  serverID := fmt.Sprintf("Server-%d", *port)

  //health check endpoint
  http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request){
    if atomic.LoadInt32(&health) == 1{
      w.WriteHeader(http.StatusOK)
      fmt.Fprintf(w, "Server %s is healthy\n", serverID)
      return 
    }

    w.WriteHeader(http.StatusServiceUnavailable)
    fmt.Fprintf(w, "Server %s is unhealthy\n", serverID)
    return 
  })

  //toggle status endpoint
    http.HandleFunc("/toggle", func(w http.ResponseWriter, r *http.Request){
      if atomic.LoadInt32(&health) == 1 {
        atomic.StoreInt32(&health, 0)
        log.Printf("Server %s is now unhealthy", serverID)
      } else {
        atomic.StoreInt32(&health, 1)
        log.Printf("Server %s is now healthy", serverID)
      }
      w.WriteHeader(http.StatusOK)
      fmt.Fprintf(w, "Toggled health status\n")
    })

  //define the main handler function
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
    log.Printf("%s received request from %s: %s %s", serverID, r.RemoteAddr, r.Method, r.URL.Path)

    //set headers
    w.Header().Set("Content-Type", "text/plain")
    w.Header().Set("Server", serverID)

    //response body
    response := fmt.Sprintf("Hello from %s!\nRequest path: %s\nRequest method: %s\nHeaders: %v\n",
    serverID,
    r.URL.Path,
    r.Method,
    r.Header)

    fmt.Fprintf(w, response)
  })
  //start the server
  addr := fmt.Sprintf(":%d", *port)
  log.Printf("%s starting on %s", serverID, addr)

  if err := http.ListenAndServe(addr, nil); err != nil {
		log.Printf("Error starting server: %v", err)
		os.Exit(1)
	}
}
