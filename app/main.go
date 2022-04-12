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

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	store := itunes.NewStore("", &http.Client{})
	srv := http.Server{
		Addr:    ":3000",
		Handler: NewApp(store),
	}

	errs, ctx := errgroup.WithContext(context.Background())

	errs.Go(func() error {
		log.Println("starting server")
		if err := srv.ListenAndServe(); err != nil {
			return err
		}
		return nil
	})

	errs.Go(func() error {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
		<-sigs

		log.Println("shutting down server")
		tc, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		return srv.Shutdown(tc)
	})

	return errs.Wait()
}
