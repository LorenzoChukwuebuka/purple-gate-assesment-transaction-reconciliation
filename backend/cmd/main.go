package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"transaction-reconciliation/models"
)

var db *gorm.DB

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("reconciliation.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}
	// Migrate the schema
	db.AutoMigrate(&models.Transaction{}, &models.Discrepancy{})
}

func main() {
	initDB()

	r := mux.NewRouter()

	// CORS handling
	corsMiddleware := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	// API endpoints
	r.HandleFunc("/api/reconcile", reconcileHandler).Methods("POST")
	r.HandleFunc("/api/discrepancies", getDiscrepanciesHandler).Methods("GET")

	// Serve static files for frontend
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public")))

	port := "7080"
	fmt.Printf("Server is running on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, corsMiddleware(r)))
}

func reconcileHandler(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form data (max 100MB)
	err := r.ParseMultipartForm(100 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get files from form
	fileA, handlerA, err := r.FormFile("fileA")
	if err != nil {
		http.Error(w, "Error retrieving file A: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer fileA.Close()

	fileB, handlerB, err := r.FormFile("fileB")
	if err != nil {
		http.Error(w, "Error retrieving file B: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer fileB.Close()

	// Create temporary files
	tempFileA, err := os.CreateTemp("", "sourceA-*.csv")
	if err != nil {
		http.Error(w, "Error creating temp file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFileA.Name())
	defer tempFileA.Close()

	tempFileB, err := os.CreateTemp("", "sourceB-*.csv")
	if err != nil {
		http.Error(w, "Error creating temp file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFileB.Name())
	defer tempFileB.Close()

	// Copy uploaded files to temp files
	_, err = io.Copy(tempFileA, fileA)
	if err != nil {
		http.Error(w, "Error saving file A: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = io.Copy(tempFileB, fileB)
	if err != nil {
		http.Error(w, "Error saving file B: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Close and reopen temp files for reading
	tempFileA.Close()
	tempFileB.Close()

	tempFileA, _ = os.Open(tempFileA.Name())
	tempFileB, _ = os.Open(tempFileB.Name())

	// Clear previous reconciliation data
	db.Exec("DELETE FROM transactions")
	db.Exec("DELETE FROM discrepancies")

	// Process files and perform reconciliation
	response, err := processFiles(tempFileA, tempFileB, handlerA.Filename, handlerB.Filename)
	if err != nil {
		http.Error(w, "Error during reconciliation: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func processFiles(fileA, fileB *os.File, filenameA, filenameB string) (*models.ReconcileResponse, error) {
	// Process and load files into DB
	transactionsA, err := loadTransactions(fileA, "A")
	if err != nil {
		return nil, fmt.Errorf("error loading file A: %v", err)
	}

	transactionsB, err := loadTransactions(fileB, "B")
	if err != nil {
		return nil, fmt.Errorf("error loading file B: %v", err)
	}

	// Perform reconciliation
	return reconcileTransactions(transactionsA, transactionsB)
}

func loadTransactions(file *os.File, source string) (map[string]*models.Transaction, error) {
	transactions := make(map[string]*models.Transaction)

	// Reset file pointer to beginning
	file.Seek(0, 0)

	// Create CSV reader
	reader := csv.NewReader(file)

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, err
	}

	// Map header columns to indices
	headerMap := make(map[string]int)
	for i, column := range header {
		headerMap[strings.TrimSpace(strings.ToLower(column))] = i
	}

	// Verify required columns exist
	requiredColumns := []string{"transaction_id", "timestamp", "amount", "currency", "status"}
	for _, col := range requiredColumns {
		if _, exists := headerMap[col]; !exists {
			return nil, fmt.Errorf("missing required column: %s", col)
		}
	}

	// Read and process each row
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// Extract fields
		transactionID := record[headerMap["transaction_id"]]
		timestamp := record[headerMap["timestamp"]]
		amountStr := record[headerMap["amount"]]
		currency := record[headerMap["currency"]]
		status := record[headerMap["status"]]

		// Parse amount
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			// Log the error but continue processing
			log.Printf("Error parsing amount for transaction %s: %v", transactionID, err)
			continue
		}

		// Create transaction object
		transaction := &models.Transaction{
			TransactionID: transactionID,
			Timestamp:     timestamp,
			Amount:        amount,
			Currency:      currency,
			Status:        status,
			Source:        source,
		}

		// Add to map
		transactions[transactionID] = transaction

		// Save to database
		db.Create(transaction)
	}

	return transactions, nil
}

func reconcileTransactions(transactionsA, transactionsB map[string]*models.Transaction) (*models.ReconcileResponse, error) {
	response := &models.ReconcileResponse{
		MissingInA:       []string{},
		MissingInB:       []string{},
		AmountMismatches: []models.AmountMismatch{},
		StatusMismatches: []models.StatusMismatch{},
		Summary: models.ReconciliationSummary{
			TotalTransactionsA: len(transactionsA),
			TotalTransactionsB: len(transactionsB),
		},
	}

	// Check for missing transactions in A
	for id, _ := range transactionsB {
		if _, exists := transactionsA[id]; !exists {
			response.MissingInA = append(response.MissingInA, id)

			// Save discrepancy to database
			discrepancy := models.Discrepancy{
				TransactionID: id,
				Type:          "MISSING_IN_A",
				Details:       fmt.Sprintf("Transaction %s found in System B but missing in System A", id),
			}
			db.Create(&discrepancy)
		}
	}
	response.Summary.MissingInACount = len(response.MissingInA)

	// Check for missing transactions in B and mismatches
	for id, txA := range transactionsA {
		txB, exists := transactionsB[id]
		if !exists {
			response.MissingInB = append(response.MissingInB, id)

			// Save discrepancy to database
			discrepancy := models.Discrepancy{
				TransactionID: id,
				Type:          "MISSING_IN_B",
				Details:       fmt.Sprintf("Transaction %s found in System A but missing in System B", id),
			}
			db.Create(&discrepancy)
			continue
		}

		// Check for amount mismatches
		if txA.Amount != txB.Amount {
			mismatch := models.AmountMismatch{
				TransactionID: id,
				AmountA:       txA.Amount,
				AmountB:       txB.Amount,
				Currency:      txA.Currency,
			}
			response.AmountMismatches = append(response.AmountMismatches, mismatch)

			// Save discrepancy to database
			discrepancy := models.Discrepancy{
				TransactionID: id,
				Type:          "AMOUNT_MISMATCH",
				Details: fmt.Sprintf("Amount mismatch for %s: A=%f %s, B=%f %s",
					id, txA.Amount, txA.Currency, txB.Amount, txB.Currency),
			}
			db.Create(&discrepancy)
		}

		// Check for status mismatches
		if txA.Status != txB.Status {
			mismatch := models.StatusMismatch{
				TransactionID: id,
				StatusA:       txA.Status,
				StatusB:       txB.Status,
			}
			response.StatusMismatches = append(response.StatusMismatches, mismatch)

			// Save discrepancy to database
			discrepancy := models.Discrepancy{
				TransactionID: id,
				Type:          "STATUS_MISMATCH",
				Details: fmt.Sprintf("Status mismatch for %s: A=%s, B=%s",
					id, txA.Status, txB.Status),
			}
			db.Create(&discrepancy)
		}
	}
	response.Summary.MissingInBCount = len(response.MissingInB)
	response.Summary.AmountMismatchesCount = len(response.AmountMismatches)
	response.Summary.StatusMismatchesCount = len(response.StatusMismatches)
	response.Summary.TotalDiscrepancies = response.Summary.MissingInACount +
		response.Summary.MissingInBCount +
		response.Summary.AmountMismatchesCount +
		response.Summary.StatusMismatchesCount

	return response, nil
}

func getDiscrepanciesHandler(w http.ResponseWriter, r *http.Request) {
	// Get pagination parameters
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("pageSize")
	discrepancyType := r.URL.Query().Get("type")

	page := 1
	pageSize := 50

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 1000 {
			pageSize = ps
		}
	}

	// Build the query
	query := db.Model(&models.Discrepancy{})
	if discrepancyType != "" {
		query = query.Where("type = ?", discrepancyType)
	}

	// Get total count
	var totalCount int64
	query.Count(&totalCount)

	// Get paginated results
	var discrepancies []models.Discrepancy
	query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&discrepancies)

	// Calculate total pages
	totalPages := int((totalCount + int64(pageSize) - 1) / int64(pageSize))

	// Create response
	response := models.PaginatedResponse{
		Page:          page,
		PageSize:      pageSize,
		TotalPages:    totalPages,
		TotalItems:    int(totalCount),
		Discrepancies: discrepancies,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
