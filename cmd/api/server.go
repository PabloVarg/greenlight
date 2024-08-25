package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	srv := http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		ErrorLog:     log.New(app.logger, "", 0),
	}

	go func() {
		done := make(chan os.Signal, 1)
		signal.Notify(done, syscall.SIGTERM, syscall.SIGINT)

		s := <-done

		app.logger.Info("caught signal", map[string]string{
			"signal": s.String(),
		})
		os.Exit(0)
	}()

	app.logger.Info("Starting %s server on %s", map[string]string{
		"env":  app.config.env,
		"addr": srv.Addr,
	})
	return srv.ListenAndServe()
}
