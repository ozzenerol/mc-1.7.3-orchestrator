package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/ozzenerol/mc-1.7.3-orchestrator/internal/config"
	"github.com/ozzenerol/mc-1.7.3-orchestrator/internal/db"
	"github.com/ozzenerol/mc-1.7.3-orchestrator/internal/host"
	"github.com/ozzenerol/mc-1.7.3-orchestrator/internal/instance"
	"github.com/pressly/goose/v3"
)

const (
	ConfigPath = "config.yaml"
)

//go:embed migrations/*.sql
var migrations embed.FS

func main() {
	ctx := context.Background()

	cfg, err := config.Load(ConfigPath)
	if err != nil {
		log.Fatalf("Could not load configuration at %s: %v", ConfigPath, err)
	}

	databaseURL, err := config.GetDatabaseURL(cfg)
	fmt.Printf("database url: %s\n", databaseURL)
	if err != nil {
		log.Fatal(err)
	}

	pool, err := pgxpool.New(ctx, databaseURL) 
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	sqlDB := stdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()

	goose.SetBaseFS(migrations)
	
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal(err)
	}

	if err := goose.Up(sqlDB, "migrations"); err != nil {
		log.Fatal(err)
	}

	q := db.New(pool)
	
	mux := http.NewServeMux()
	host.NewHandler(host.NewService(q)).Register(mux)
	instance.NewHandler(instance.NewService(q)).Register(mux)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
