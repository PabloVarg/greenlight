package main

import (
	"context"
	"database/sql"
	"expvar"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"greenlight.pvargasb.com/internal/data"
	"greenlight.pvargasb.com/internal/jsonlog"
	"greenlight.pvargasb.com/internal/mailer"
)

var (
	version   string
	buildTime string
)

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	cors struct {
		trustedOrigins []string
	}
}

type application struct {
	wg     sync.WaitGroup
	models data.Models
	config config
	logger *jsonlog.Logger
	mailer mailer.Mailer
}

func main() {
	var config config
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
		"",
		"PostgreSQL DNS",
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
	flag.Float64Var(
		&config.limiter.rps,
		"limiter-rps",
		2,
		"Rate limiter maximum requests per second",
	)
	flag.IntVar(
		&config.limiter.burst,
		"limiter-burst",
		4,
		"Rate limiter maximum burst",
	)
	flag.BoolVar(&config.limiter.enabled,
		"limiter-enabled",
		true,
		"Enable rate limiter",
	)
	flag.StringVar(&config.smtp.host,
		"smtp-host",
		"localhost",
		"SMTP host",
	)
	flag.IntVar(&config.smtp.port,
		"smtp-port",
		1025,
		"SMTP port",
	)
	flag.StringVar(&config.smtp.username,
		"smtp-username",
		"mailpit",
		"SMTP username",
	)
	flag.StringVar(
		&config.smtp.password,
		"smtp-password",
		"mailpit",
		"SMTP password",
	)
	flag.StringVar(
		&config.smtp.sender,
		"smtp-sender",
		"Greenlight <no-reply@greenlight.pvargasb.net>",
		"SMTP sender",
	)
	flag.Func(
		"cors-trusted-origins",
		"Trusted CORS origins (space separated)",
		func(s string) error {
			config.cors.trustedOrigins = strings.Fields(s)
			return nil
		},
	)
	displayVersion := flag.Bool(
		"version",
		false,
		"Display version and exit",
	)
	flag.Parse()

	if *displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		fmt.Printf("Build time:\t%s\n", buildTime)
		return
	}

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := openDB(config)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	logger.Info("Connection to DB(%s) established\n", map[string]string{
		"dsn": config.db.dsn,
	})

	expvar.NewString("version").Set(version)
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))
	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))
	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))

	app := application{
		models: *data.NewModels(db),
		config: config,
		logger: logger,
		mailer: mailer.New(config.smtp.host, config.smtp.port, config.smtp.username, config.smtp.password, config.smtp.sender),
	}

	if err := app.serve(); err != nil {
		logger.Fatal(err, nil)
	}
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
