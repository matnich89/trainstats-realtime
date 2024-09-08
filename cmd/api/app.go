package cmd

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/matnich89/trainstats-realtime/handler/national"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	router          *chi.Mux
	nationalHandler *national.Handler
}

func NewApp(router *chi.Mux, handler *national.Handler) *App {
	return &App{
		router:          router,
		nationalHandler: handler,
	}
}

func (a *App) routes() {
	a.router.Get("/national", a.nationalHandler.HandleNationalData)
}

func (a *App) Serve() error {

	a.nationalHandler.Listen()

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
