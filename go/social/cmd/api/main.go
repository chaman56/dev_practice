package main

import (
	"log"
	"time"

	"github.com/chaman56/dev_practice/go/social/internal/db"
	"github.com/chaman56/dev_practice/go/social/internal/env"
	"github.com/chaman56/dev_practice/go/social/internal/store"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found: %v", err)
	}

	cfg := Config{
		addr: env.GetEnv("ADDR", ":8080"),
		db: dbConfig{
			addr:         env.GetEnv("DATABASE_URL", "postgres://admin:adminpassword@localhost/social?sslmode=disable"),
			maxOpenConns: env.GetEnvAsInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetEnvAsInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetEnvAsDuration("DB_MAX_IDLE_TIME", 15*time.Minute),
		},
	}

	db, err := db.NewDB(
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()
	log.Println("Databse connection established.")

	store := store.NewStorage(db)

	app := &application{
		config: cfg,
		store:  store,
	}
	mux := app.mount()
	log.Fatal(app.run(mux))
}
