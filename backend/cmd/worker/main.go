package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fifriz/internal/config"
	"fifriz/internal/database"
	"fifriz/internal/server"
)

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

	_, documentWorker, integrationService := server.New(cfg, db)

	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()

	go documentWorker.Run(workerCtx)
	go integrationService.RunSyncScheduler(workerCtx)

	log.Printf("background worker process started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	workerCancel()
	time.Sleep(200 * time.Millisecond)

	sqlDB, err := db.DB()
	if err == nil {
		_ = sqlDB.Close()
	}
}
