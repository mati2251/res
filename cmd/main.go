package main

import (
	"fmt"
	"log"
	"net/http"
	handlers "res/pkg/http"
)

func main() {
  mux := http.NewServeMux()
  mux.HandleFunc("/job", handlers.PostJob)
  loggedMux := handlers.CommonLogger(mux)
	port := 8080
	log.Printf("Server started on port %d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), loggedMux))
}
