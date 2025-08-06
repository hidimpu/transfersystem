package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"

	"transfersystem/internal/api"
	"transfersystem/internal/config"
	"transfersystem/internal/db"
)

func main() {
	_ = godotenv.Load(".env")
	cfg := config.LoadConfig()
	dbConn := db.NewPostgresDB(cfg)
	handler := api.NewHandler(dbConn)

	r := chi.NewRouter()
	r.Post("/accounts", handler.HandleAccounts)
	r.Get("/accounts/{account_id}", handler.HandleGetAccount)
	r.Post("/transactions", handler.HandleTransactions)
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Printf("Server started on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Graceful shutdown failed: %v", err)
	}
	log.Println("Server exited properly")
}
