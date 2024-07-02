package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"greenlight.pvargasb.com/internal/data"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
}

type application struct {
	models data.Models
	config config
	logger *log.Logger
}

func main() {
	var config config
	defaultDSN, ok := os.LookupEnv("POSTGRES_DSN")
	if !ok {
		defaultDSN = "postgres://greenlight:greenlight@localhost/greenlight?sslmode=disable"
	}

	flag.IntVar(
		&config.port,
		"port",
		4000,
		"API server port",
	)
	flag.StringVar(
		&config.env,
		"env",
		"development",
		"Environment (development|staging|production)",
	)
	flag.StringVar(
		&config.db.dsn,
		"dsn",
		defaultDSN,
		fmt.Sprintf("PostgreSQL DNS (default: %s)", defaultDSN),
	)
	flag.IntVar(
		&config.db.maxOpenConns,
		"db-max-open-conns",
		25,
		"Postgres max open connections",
	)
	flag.IntVar(
		&config.db.maxIdleConns,
		"db-max-idle-conns",
		25,
		"Postgres max idle connections",
	)
	flag.StringVar(
		&config.db.maxIdleTime,
		"db-max-idle-time",
		"15m",
		"Postgres max idle time",
	)
	flag.Parse()

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	db, err := openDB(config)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	logger.Printf("Connection to DB(%s) established\n", config.db.dsn)

	app := application{
		models: *data.NewModels(db),
		config: config,
		logger: logger,
	}

	srv := http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.Printf("Starting %s server on %s", app.config.env, srv.Addr)
	err = srv.ListenAndServe()
	logger.Fatal(err)
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	maxIdleTime, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		log.Fatal(err)
	}
	db.SetConnMaxIdleTime(maxIdleTime)
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
