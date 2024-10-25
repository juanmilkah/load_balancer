package main

import (
  "fmt"
  "os"
  "net/http"
  "flag"
  "log"
)

func main(){
  //define commandline flags for port numbers
  port := flag.Int("port", 8081, "port to server on")
  flag.Parse()

  //create a uniques server identifier
  serverID := fmt.Sprintf("Server-%d", *port)

  //define the handler function
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
