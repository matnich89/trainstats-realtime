package main

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/matnich89/network-rail-client/client"
	cmd "github.com/matnich89/trainstats-service-template/cmd/api"
	"github.com/matnich89/trainstats-service-template/handler"
	"log"
	"os"
)

func main() {
	router := chi.NewMux()

	username := os.Getenv("NR_USERNAME")
	password := os.Getenv("NR_PASSWORD")

	ctx := context.Background()

	nrClient, err := client.NewNetworkRailClient(ctx, username, password)

	if err != nil {
		log.Fatalln(fmt.Errorf("error creating network rail client: %w", err))
	}

	requestHandler, err := handler.NewHandler(nrClient)

	if err != nil {
		log.Fatalln(fmt.Errorf("error creating handler: %w", err))
	}

	app := cmd.NewApp(router, requestHandler)

	err = app.Serve()

	if err != nil {
		log.Fatalln(err)
	}

}
