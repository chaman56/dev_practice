package main

import (
	"log"

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
	}

	store := store.NewStorage(nil)

	app := &application{
		config: cfg,
		store:  store,
	}
	mux := app.mount()
	log.Fatal(app.run(mux))
}
