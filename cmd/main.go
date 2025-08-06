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
	"github.com/hidimpu/transfersystem/internal/repository"
	"github.com/hidimpu/transfersystem/internal/service"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	dbConn, err := db.InitDB()
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}
	log.Println("DB Connection Established!")
	defer dbConn.Close()

	// Initialize repositories (Data Access Layer)
	accountRepo := repository.NewAccountRepository(dbConn)
	transactionRepo := repository.NewTransactionRepository(dbConn)

	// Initialize services (Business Logic Layer)
	accountService := service.NewAccountService(accountRepo)
	transactionService := service.NewTransactionService(dbConn, accountRepo, transactionRepo)

	// Initialize handlers (Controller Layer)
	transactionHandler := api.NewTransactionHandler(transactionService)

	// Setup router (View Layer)
	r := chi.NewRouter()

	// Account routes
	r.Route("/accounts", func(r chi.Router) {
		r.Post("/", api.CreateAccountServiceHandler(accountService))
		r.Get("/{account_id}", api.GetAccountServiceHandler(accountService))
	})

	// Transaction routes
	r.Route("/transactions", func(r chi.Router) {
		r.Post("/", transactionHandler.TransferFunds)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 Server started on port: %s\n", port)
	fmt.Println("📊 Database locks: FOR UPDATE with Serializable isolation")
	fmt.Println("🏗️  Architecture: MVC with proper separation of concerns")
	fmt.Println("🔒 Concurrency: Row-level locking with atomic transactions")

	http.ListenAndServe(":"+port, r)
}
