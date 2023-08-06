package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kickbu2towski/brb-api/internal/data"
)

type application struct {
	logger *log.Logger
	config config
	pool   *pgxpool.Pool
	models *data.Models
	hub    *Hub
}

type config struct {
	port   string
	dsn    string
	webURL string
	cors   struct {
		allowedOrigins []string
	}
	google struct {
		clientID     string
		clientSecret string
		redirectURL  string
	}
}

func main() {
	var cfg config
	parseFlags(&cfg)

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	pool, err := getPool(context.Background(), cfg.dsn)
	if err != nil {
		logger.Fatal(err)
	}

	models := data.NewModels(pool)
	app := application{
		logger: logger,
		config: cfg,
		pool:   pool,
		models: models,
		hub:    NewHub(models),
	}

	server := &http.Server{
		Handler:      app.routes(),
		Addr:         fmt.Sprintf(":%s", cfg.port),
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	go app.hub.run()
	logger.Printf("server starting at port %s", cfg.port)
	err = server.ListenAndServe()
	logger.Fatal(err)
}

func getPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	err = pool.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func parseFlags(cfg *config) {
	flag.StringVar(&cfg.port, "port", "6969", "API server port")
	flag.StringVar(&cfg.dsn, "dsn", os.Getenv("POSTGRES_DSN"), "PostgreSQL DSN")
	flag.StringVar(&cfg.webURL, "web-url", "http://localhost:3000", "Frontend URL")

	flag.StringVar(&cfg.google.clientID, "google-client-id", os.Getenv("GOOGLE_CLIENT_ID"), "Google Client ID")
	flag.StringVar(&cfg.google.clientSecret, "google-cient-secret", os.Getenv("GOOGLE_CLIENT_SECRET"), "Google Client Secret")
	flag.StringVar(&cfg.google.redirectURL, "google-redirect-url", os.Getenv("GOOGLE_REDIRECT_URL"), "Google Redirect URL")

	cfg.cors.allowedOrigins = []string{"http://localhost:3000"}
	flag.Func("allowed-origins", "A list of allowed origins", func(s string) error {
		cfg.cors.allowedOrigins = strings.Split(s, " ")
		return nil
	})

	flag.Parse()
}
