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
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	store := itunes.NewStore("", &http.Client{Timeout: 2 * time.Second})
	limiter := rate.NewLimiter(rate.Every(time.Minute), 20)
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	app, err := NewApp(store, true, limiter, infoLog, errorLog)
	if err != nil {
		return err
	}

	srv := http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           app,
		ErrorLog:          errorLog,
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       1 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	errs, ctx := errgroup.WithContext(context.Background())

	errs.Go(func() error {
		infoLog.Printf("starting server: %d\n", port)
		return srv.ListenAndServe()
	})

	errs.Go(func() error {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
		<-sigs

		infoLog.Printf("shutting down server: %d\n", port)
		tc, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		return srv.Shutdown(tc)
	})

	return errs.Wait()
}
