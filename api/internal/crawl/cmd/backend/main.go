package main

import (
	"context"
	"log"
	"os"
	server "sigs.k8s.io/kustomize/api/internal/crawl/backend"
	"strconv"
)

func main() {
	portStr := os.Getenv("PORT")
	port, err := strconv.Atoi(portStr)
	if portStr == "" || err != nil {
		log.Fatalf("$PORT(%s) must be set to an integer\n", portStr)
	}

	ctx := context.Background()

	ks, err := server.NewKustomizeSearch(ctx)
	if err != nil {
		log.Fatalf("Error creating kustomize server: %v", ks)
	}

	err = ks.Serve(port)
	if err != nil {
		log.Fatalf("Error while running server: %v", err)
	}
}
