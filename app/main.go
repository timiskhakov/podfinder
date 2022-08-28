package main

import (
	"context"
	"fmt"
	"github.com/timiskhakov/podfinder/app/itunes"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const port = 3000

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	store := itunes.NewStore("", &http.Client{})
	srv := http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           NewApp(store),
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       1 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	errs, ctx := errgroup.WithContext(context.Background())

	errs.Go(func() error {
		log.Printf("starting server: %d\n", port)
		return srv.ListenAndServe()
	})

	errs.Go(func() error {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
		<-sigs

		log.Printf("shutting down server: %d\n", port)
		tc, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		return srv.Shutdown(tc)
	})

	return errs.Wait()
}
