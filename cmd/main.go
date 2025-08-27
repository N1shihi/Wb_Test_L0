package main

import (
	"Wb_Test_L0/internal/config"
	"Wb_Test_L0/internal/handler"
	"Wb_Test_L0/internal/kafka"
	"Wb_Test_L0/internal/repository"
	"Wb_Test_L0/internal/server"
	"Wb_Test_L0/internal/service"
	"Wb_Test_L0/internal/storage"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.MustLoad()
	fmt.Println("Loaded config:", cfg)

	db, err := storage.New(cfg)
	if err != nil {
		log.Fatalf("storage.New error: %v", err)
	}
	defer db.Close()

	repo := repository.NewRepository(db)
	svc := service.NewService(repo, cfg)
	h := handler.New(svc)
	cons, err := kafka.NewConsumer(h, cfg)
	if err != nil {
		log.Fatalf("kafka.NewConsumer error: %v", err)
	}

	go cons.Start()

	srv := server.New(cfg, db)
	go func() {
		if err := srv.Run(h); err != nil {
			log.Printf("server run error: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Shutdown signal received")

	if err := cons.Stop(); err != nil {
		log.Printf("Error stopping consumer: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Stop(ctx); err != nil {
		log.Printf("Error stopping server: %v", err)
	}

	log.Println("Shutdown complete")
}
