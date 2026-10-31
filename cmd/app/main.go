package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/VsemoGosi-Adventure/events/internal/config"
	"github.com/VsemoGosi-Adventure/events/internal/consumer"
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
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
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
	eventService := service.NewEventService(db)
	eventHandler := v1.NewEventHandler(eventService)
	eventConsumer := consumer.NewConsumer(cfg.ConsumerConfig, db)
	defer func() {
		if err := eventConsumer.Close(); err != nil {
			slog.Error("close event consumer failed", "error", err)
		}
	}()
	go func() {
		if err := eventConsumer.Run(ctx); err != nil {
			slog.Error("event consumer stopped", "error", err)
			stop()
		}
	}()

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.SlogLogger())

	eventHandler.RegisterRoutes(router)

	port := "8080"
	slog.Info("starting events-service", slog.String("port", port))
	server := &http.Server{Addr: ":" + port, Handler: router}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("server shutdown failed", "error", err)
		}
	}()
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server stopped with error", slog.Any("error", err))
		log.Fatalf("server stopped: %v", err)
	}
}
