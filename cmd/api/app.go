package cmd

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/matnich89/trainstats-realtime/handler/national"
	"github.com/matnich89/trainstats-realtime/service"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type App struct {
	router             *chi.Mux
	nationalHandler    *national.Handler
	networkRailService *service.NetworkRail
	wg                 *sync.WaitGroup
	shutdownCh         chan struct{}
}

func NewApp(router *chi.Mux, nationalHandler *national.Handler, nrService *service.NetworkRail) *App {
	return &App{
		router:             router,
		nationalHandler:    nationalHandler,
		networkRailService: nrService,
		wg:                 &sync.WaitGroup{},
		shutdownCh:         make(chan struct{}),
	}
}

func (a *App) routes() {
	a.router.Get("/passenger", a.nationalHandler.HandlePassengerData)
}

func (a *App) Serve(ctx context.Context) error {
	a.wg.Add(2)
	go func() {
		defer a.wg.Done()
		a.nationalHandler.Listen(a.shutdownCh)
	}()

	go func() {
		defer a.wg.Done()
		a.networkRailService.ProcessData(a.shutdownCh)
	}()

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      a.router,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	shutdownError := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit
		log.Printf("caught signal %s", s.String())

		close(a.shutdownCh)

		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		shutdownError <- srv.Shutdown(ctx)
	}()

	log.Println("starting api...")
	a.routes()
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	log.Println("waiting for goroutines to finish...")
	a.wg.Wait()
	log.Println("stopped api")

	return nil
}
