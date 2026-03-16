package utils

import (
	"log"
	"os"
	"time"
)

// Logger provides structured logging functionality
type Logger struct {
	*log.Logger
}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

// LogError logs an error with context
func (l *Logger) LogError(operation, errorType, message string, err error) {
	l.Printf("[ERROR] %s | Type: %s | Message: %s | Error: %v | Time: %s",
		operation, errorType, message, err, time.Now().Format(time.RFC3339))
}

// LogInfo logs an informational message
func (l *Logger) LogInfo(operation, message string) {
	l.Printf("[INFO] %s | Message: %s | Time: %s",
		operation, message, time.Now().Format(time.RFC3339))
}

// LogWarning logs a warning message
func (l *Logger) LogWarning(operation, message string) {
	l.Printf("[WARN] %s | Message: %s | Time: %s",
		operation, message, time.Now().Format(time.RFC3339))
}

// LogTransfer logs transfer-specific information
func (l *Logger) LogTransfer(operation string, sourceID, destID int64, amount string, success bool) {
	status := "SUCCESS"
	if !success {
		status = "FAILED"
	}
	l.Printf("[TRANSFER] %s | Source: %d | Destination: %d | Amount: %s | Status: %s | Time: %s",
		operation, sourceID, destID, amount, status, time.Now().Format(time.RFC3339))
}

// LogAccount logs account-specific information
func (l *Logger) LogAccount(operation string, accountID int64, balance string, success bool) {
	status := "SUCCESS"
	if !success {
		status = "FAILED"
	}
	l.Printf("[ACCOUNT] %s | AccountID: %d | Balance: %s | Status: %s | Time: %s",
		operation, accountID, balance, status, time.Now().Format(time.RFC3339))
}

// Global logger instance
var GlobalLogger = NewLogger()
