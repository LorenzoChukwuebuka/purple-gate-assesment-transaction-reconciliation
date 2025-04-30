package models
import (
	"gorm.io/gorm"
)

// Transaction represents a financial transaction
type Transaction struct {
	gorm.Model
	TransactionID string  `json:"transaction_id" gorm:"index"`
	Timestamp     string  `json:"timestamp"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	Status        string  `json:"status"`
	Source        string  `json:"source"` // "A" or "B"
}

// Discrepancy represents a reconciliation discrepancy
type Discrepancy struct {
	gorm.Model
	TransactionID string `json:"transaction_id"`
	Type          string `json:"type"` // "MISSING_IN_A", "MISSING_IN_B", "AMOUNT_MISMATCH", "STATUS_MISMATCH"
	Details       string `json:"details"`
}

// ReconcileResponse represents the API response
type ReconcileResponse struct {
	MissingInA      []string               `json:"missing_in_a"`
	MissingInB      []string               `json:"missing_in_b"`
	AmountMismatches []AmountMismatch      `json:"amount_mismatches"`
	StatusMismatches []StatusMismatch      `json:"status_mismatches"`
	Summary         ReconciliationSummary  `json:"summary"`
}

// AmountMismatch represents a transaction with amount discrepancy
type AmountMismatch struct {
	TransactionID string  `json:"transaction_id"`
	AmountA       float64 `json:"amount_a"`
	AmountB       float64 `json:"amount_b"`
	Currency      string  `json:"currency"`
}

// StatusMismatch represents a transaction with status discrepancy
type StatusMismatch struct {
	TransactionID string `json:"transaction_id"`
	StatusA       string `json:"status_a"`
	StatusB       string `json:"status_b"`
}

// ReconciliationSummary provides summary statistics
type ReconciliationSummary struct {
	TotalTransactionsA    int `json:"total_transactions_a"`
	TotalTransactionsB    int `json:"total_transactions_b"`
	MissingInACount       int `json:"missing_in_a_count"`
	MissingInBCount       int `json:"missing_in_b_count"`
	AmountMismatchesCount int `json:"amount_mismatches_count"`
	StatusMismatchesCount int `json:"status_mismatches_count"`
	TotalDiscrepancies    int `json:"total_discrepancies"`
}

// PaginatedResponse for handling large result sets
type PaginatedResponse struct {
	Page          int         `json:"page"`
	PageSize      int         `json:"page_size"`
	TotalPages    int         `json:"total_pages"`
	TotalItems    int         `json:"total_items"`
	Discrepancies interface{} `json:"discrepancies"`
}
