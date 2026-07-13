package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"fifriz/internal/config"
	"fifriz/internal/database"
	"fifriz/internal/server"
)

func runInProcessWorker() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv("RUN_IN_PROCESS_WORKER")))
	if value == "false" || value == "0" {
		return false
	}
	return true
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := database.ConnectWithRetry(cfg.Database, 30, 2*time.Second)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}

	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	app, documentWorker, integrationService := server.New(cfg, db)

	var workerCancel context.CancelFunc
	if runInProcessWorker() {
		var workerCtx context.Context
		workerCtx, workerCancel = context.WithCancel(context.Background())
		go documentWorker.Run(workerCtx)
		go integrationService.RunSyncScheduler(workerCtx)
		log.Printf("in-process background worker enabled")
	} else {
		log.Printf("in-process background worker disabled (use separate worker process)")
	}

	go func() {
		if err := app.Listen(cfg.HTTP.Addr()); err != nil {
			log.Printf("http server stopped: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	if workerCancel != nil {
		workerCancel()
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}

	sqlDB, err := db.DB()
	if err == nil {
		_ = sqlDB.Close()
	}

	time.Sleep(100 * time.Millisecond)
}
