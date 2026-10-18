package main

import (
	"context"
	"log"

	"github.com/VsemoGosi-Adventure/events/internal/config"
	"github.com/VsemoGosi-Adventure/events/internal/storage"
	"github.com/VsemoGosi-Adventure/events/internal/storage/postgres"
)

func main() {
	ctx := context.Background()
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	pool, err := postgres.NewPostgresPool(ctx, cfg.PostgresConfig)
	if err != nil {
		log.Fatal(err)
	}

	db := storage.NewStorage(pool)
	defer func() {
		if err := db.Close(); err != nil {
			log.Println("Error of close database connection", err)
		}
	}()

	log.Println("Connect succsessful")
}
