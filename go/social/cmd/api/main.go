package main

import (
	"log"

	"github.com/chaman56/dev_practice/go/social/internal/env"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found: %v", err)
	}

	cfg := Config{
		addr: env.GetEnv("ADDR", ":8080"),
	}

	app := &application{
		config: cfg,
	}
	mux := app.mount()
	log.Fatal(app.run(mux))
}
