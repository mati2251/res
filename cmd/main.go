package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"res/pkg/config"
	handlers "res/pkg/http"
)

func main() {
	// get config file path from command line
	configFile := flag.String("config", ".config.yaml", "config file path")
	flag.Parse()
	conf, err := config.Set(*configFile)
	if err != nil {
		log.Fatalf("failed to set config: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/job", handlers.PostJob)
	mux.HandleFunc("/job/{id}", handlers.Job)
	loggedMux := handlers.CommonLogger(mux)

	port := conf.Port
	log.Printf("Server started on port %d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), loggedMux))
}
