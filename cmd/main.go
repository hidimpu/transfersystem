package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"

	"github.com/hidimpu/transfersystem/internal/api"
	"github.com/hidimpu/transfersystem/internal/db"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	dbConn, err := db.InitDB()
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}
	defer dbConn.Close()

	r := chi.NewRouter()

	r.Route("/accounts", func(r chi.Router) {
		r.Post("/", api.CreateAccountHandler(dbConn))
		r.Get("/{account_id}", api.GetAccountHandler(dbConn))
	})

	r.Post("/transactions", api.CreateTransactionHandler(dbConn))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Server started on port:", port)
	http.ListenAndServe(":"+port, r)
}
