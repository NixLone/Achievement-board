package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"firegoals/internal/auth"
	"firegoals/internal/config"
	"firegoals/internal/db"
	api "firegoals/internal/http"
	"firegoals/internal/repo"
	"firegoals/internal/service"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}
	defer pool.Close()

	authManager := auth.NewManager(cfg.JWTSecret)
	repository := repo.New(pool)
	svc := service.New(repository, authManager)

	handler := &api.API{Repo: repository, Service: svc, Auth: authManager}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler.Router(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("server listening on %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
}
