package main

import (
	"context"
	"fmt"
	"github.com/timiskhakov/podfinder/app/itunes"
	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const port = 3000

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100
	t.MaxIdleConnsPerHost = 100

	app, err := NewApp(&AppConfig{
		Store:            itunes.NewStore("", &http.Client{Timeout: 2 * time.Second, Transport: t}),
		IsLimiterEnabled: true,
		Limiter:          rate.NewLimiter(rate.Every(time.Minute), 20),
	})
	if err != nil {
		return err
	}

	srv := http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           app,
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       1 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	errs, ctx := errgroup.WithContext(ctx)
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
