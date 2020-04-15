package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/DoubleDi/golang_projects/transactions/pkg/transaction"

	"github.com/DoubleDi/golang_projects/transactions/pkg/app"
	"github.com/DoubleDi/golang_projects/transactions/pkg/config"
	"github.com/DoubleDi/golang_projects/transactions/pkg/database"

	"github.com/chapsuk/grace"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
)

var (
	appName   string = "tx"
	buildHash string = "_dev"
	buildTime string = "_dev"
)

func main() {
	log.Printf("App: %s, Commit: %s, Build: %s", appName, buildHash, buildTime)
	rand.Seed(time.Now().UnixNano())
	ctx := context.Background()

	cfg := config.Get()
	if err := cfg.Load(appName); err != nil {
		log.Fatalf("Can't read config: %v", err)
	}
	log.Printf("Starting with config: %#v", cfg)

	db, err := sqlx.Open("pgx", fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	))
	if err != nil {
		log.Fatalf("Can't connect to db, err: %v", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalf("Can't ping db, err: %v", err)
	}

	validator := app.NewValidator()
	repo := database.NewTransactionPostgresRepository(db)

	worker := transaction.NewWorker(cfg, repo)
	go worker.Run()

	r := app.Router(repo, validator)
	stopCtx := grace.StopContext(ctx)
	go func() {
		<-stopCtx.Done()
		log.Println("Shutting down gracefully")
		worker.Close()
	}()
	log.Printf("Listening server at %s", cfg.HTTPPort)
	http.ListenAndServe(":"+cfg.HTTPPort, r)
}
