package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/jackc/pgx/v5"
	"log"
	"net/http"
	"res/pkg/config"
	"res/pkg/db"
	handlers "res/pkg/http"
)

func main() {
	configPath := flag.String("config", ".config.yaml", "config file path")
	flag.Parse()
	conf, err := config.Get(*configPath)
	if err != nil {
		log.Fatalf("failed to set config: %v", err)
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, conf.DbUrl)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)
	queries := db.New(conn)

	defer conn.Close(ctx)
	httpService := handlers.NewService(queries, &conf)

	mux := http.NewServeMux()
	mux.HandleFunc("/job", httpService.PostJob)
	mux.HandleFunc("/job/{id}", httpService.Job)
	loggedMux := handlers.CommonLogger(mux)

	port := conf.Port
	log.Printf("Server started on port %d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), loggedMux))
}
