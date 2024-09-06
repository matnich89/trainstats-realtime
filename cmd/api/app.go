package cmd

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/matnich89/trainstats-service-template/handler"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	router  *chi.Mux
	handler *handler.Handler
}

func NewApp(router *chi.Mux, handler *handler.Handler) *App {
	return &App{
		router:  router,
		handler: handler,
	}
}

func (a *App) routes() {
	a.router.Get("/trains", a.handler.HandleNationalData)
}

func (a *App) Serve() error {

	a.handler.Listen()

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
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
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
	log.Println("stopped api")

	return nil
}
