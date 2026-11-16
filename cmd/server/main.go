// Command server runs the HTTP service until SIGINT or SIGTERM.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go-service-template/internal/config"
	"go-service-template/internal/database"
	"go-service-template/internal/handler"
	"go-service-template/internal/repository"
	"go-service-template/internal/router"
	"go-service-template/internal/server"
	"go-service-template/internal/service"
	"go-service-template/pkg/logger"
)

// version is stamped into /healthz. Override at build time with -ldflags.
var version = "1.0.0"

func main() {
	cfg := config.Load()
	log := logger.New(cfg.LogLevel)

	if err := cfg.Validate(); err != nil {
		log.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	db, err := database.Open(cfg.DatabasePath)
	if err != nil {
		log.Error("database unavailable", "error", err)
		os.Exit(1)
	}
	if cfg.SeedDemo {
		if err := database.Seed(db); err != nil {
			log.Error("seeding failed", "error", err)
			os.Exit(1)
		}
	}

	store := repository.NewItemStore(db)
	items := service.NewItems(store)
	itemHandler := handler.NewItemHandler(items, log)
	engine := router.New(itemHandler, log, version)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := server.New(cfg.Addr, engine, log, cfg.ShutdownTimeout)
	if err := srv.Run(ctx); err != nil {
		log.Error("server stopped with an error", "error", err)
		os.Exit(1)
	}
	log.Info("shutdown complete")
}
