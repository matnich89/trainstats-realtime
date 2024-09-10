package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/matnich89/network-rail-client/client"
	cmd "github.com/matnich89/trainstats-realtime/cmd/api"
	"github.com/matnich89/trainstats-realtime/handler/national"
	"github.com/matnich89/trainstats-realtime/handler/traincompany"
	"github.com/matnich89/trainstats-realtime/service"
	"log"
	"os"
)

func main() {
	router := chi.NewMux()

	username := os.Getenv("NR_USERNAME")
	password := os.Getenv("NR_PASSWORD")

	ctx := context.Background()

	nrClient, err := client.NewNetworkRailClient(ctx, username, password)

	networkRailService, err := service.NewNetworkRail(nrClient)

	if err != nil {
		log.Fatal(err)
	}

	nationalHandler := national.NewHandler(networkRailService.NationalChan)
	trainOperatorHandler := traincompany.NewHandler(networkRailService.TrainOperatorChan)

	app := cmd.NewApp(router, nationalHandler, trainOperatorHandler, networkRailService)

	err = app.Serve(ctx)

	if err != nil {
		log.Fatalln(err)
	}

}
