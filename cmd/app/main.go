package main

import (
	"context"
	"log"
	"log/slog"

	"github.com/VsemoGosi-Adventure/events/internal/config"
	"github.com/VsemoGosi-Adventure/events/internal/delivery/http/middleware"
	v1 "github.com/VsemoGosi-Adventure/events/internal/delivery/http/v1"
	"github.com/VsemoGosi-Adventure/events/internal/service"
	"github.com/VsemoGosi-Adventure/events/internal/storage"
	"github.com/VsemoGosi-Adventure/events/internal/storage/postgres"
	"github.com/gin-gonic/gin"
)

// @title          Avantyura Events Service API
// @version        1.0
// @description    Микросервис управления событиями и гео-поиском.
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
	defer db.Close()

	eventService := service.NewEventService(db)
	eventHandler := v1.NewEventHandler(eventService)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.SlogLogger())

	eventHandler.RegisterRoutes(router)

	port := "8080"
	slog.Info("starting events-service", slog.String("port", port))
	if err := router.Run(":" + port); err != nil {
		slog.Error("server stopped with error", slog.Any("error", err))
		log.Fatalf("server stopped: %v", err)
	}

	log.Println("Connect succsessful")
}
