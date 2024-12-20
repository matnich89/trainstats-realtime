package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/matnich89/network-rail-client/client"
	cmd "github.com/matnich89/trainstats-realtime/cmd/api"
	"github.com/matnich89/trainstats-realtime/handler/national"
	"github.com/matnich89/trainstats-realtime/service"
	"log"
	"os"
)

func main() {
	router := chi.NewMux()

	username := os.Getenv("NR_USERNAME")
	password := os.Getenv("NR_PASSWORD")

	if username == "" || password == "" {
		log.Fatal("Missing required environment variables NR_USERNAME, NR_PASSWORD")
	}

	ctx := context.Background()

	nrClient, err := client.NewNetworkRailClient(ctx, username, password)

	if err != nil {
		log.Fatal(err)
	}

	networkRailService, err := service.NewNetworkRail(nrClient)

	if err != nil {
		log.Fatal(err)
	}

	nationalHandler := national.NewHandler(networkRailService.OperatorChan)

	app := cmd.NewApp(router, nationalHandler, networkRailService)

	err = app.Serve(ctx)

	if err != nil {
		log.Fatalln(err)
	}

}
