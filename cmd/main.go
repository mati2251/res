package main

import (
	"log"
	"net/http"
	handlers "res/pkg/http"
)

func main() {
	http.HandleFunc("/job", handlers.PostJob)
  log.Println("Server started on port 8080")
  log.Fatal(http.ListenAndServe(":8080", nil))
}
