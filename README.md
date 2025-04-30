# Transaction Ledger Reconciliation System

A system designed to reconcile and report discrepancies in financial transaction records between two systems. This solution efficiently processes large CSV files (1M+ rows), detects various discrepancies, and provides a user-friendly interface to view and analyze the results.

## Features

- **Efficient CSV Processing**: Handles large files (1M+ rows) with streaming and optimized data structures
- **Discrepancy Detection**:
  - Transactions present in System A but missing in System B (and vice versa)
  - Amount mismatches between systems
  - Status mismatches between systems
- **User-Friendly Interface**:
  - Upload CSVs through a simple interface
  - View discrepancy summary statistics
  - Browse detailed discrepancy reports with pagination
- **RESTful API**:
  - `/api/reconcile` endpoint for reconciliation
  - `/api/discrepancies` endpoint for retrieving paginated discrepancy data

## Technology Stack

- **Backend**:
  - Go with Gorilla Mux for API routing
  - GORM ORM with SQLite for persistent storage
  - Efficient CSV parsing with Go's standard library
  
- **Frontend**:
  - Vue.js for reactive UI components
  - Modern, responsive design
  - Pagination for handling large result sets

## Getting Started

### Prerequisites

- Go 1.16+
- Node.js 14+
- npm or yarn

### Backend Setup

1. Clone this repository:
   ```bash
   git clone https://github.com/LorenzoChukwuebuka/purple-gate-assesment-transaction-reconciliation.git
   cd into the project directory
   ```

2. Install Go dependencies:
   ```bash
   cd backend
   go mod tidy
   ```

3. Run the backend server:
   ```bash
   cd backend
   go run main.go
   ```
   The server will start at http://localhost:7080

### Frontend Setup

1. Navigate to the frontend directory:
   ```bash
   cd frontend
   ```

2. Install dependencies:
   ```bash
   npm install
   # or
   yarn install
   ```

3. Start the development server:
   ```bash
   npm run serve
   # or
   yarn serve
   ```
   The frontend will be available at http://localhost:5173

### Generate Test Data

You can generate sample test data using the provided script:
A test data has already been generated you can use it to test the system.

```bash
cd testdata
go run generate_testdata_data.go
```

This will create two CSV files with controlled discrepancies for testing.

### Backend (Go)

#### Data Models

```go
 
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

```

## API Endpoints

### POST /api/reconcile

Reconciles two CSV files and returns discrepancies.

**Request**:
- Multipart form data with two files: `fileA` and `fileB`

**Response**:
```json
{
  "missing_in_a": ["TXN-00001234", "TXN-00005678"],
  "missing_in_b": ["TXN-00002345", "TXN-00006789"],
  "amount_mismatches": [
    {
      "transaction_id": "TXN-00003456",
      "amount_a": 123.45,
      "amount_b": 123.50,
      "currency": "USD"
    }
  ],
  "status_mismatches": [
    {
      "transaction_id": "TXN-00004567",
      "status_a": "SUCCESS",
      "status_b": "FAILED"
    }
  ],
  "summary": {
    "total_transactions_a": 9800,
    "total_transactions_b": 9850,
    "missing_in_a_count": 200,
    "missing_in_b_count": 150,
    "amount_mismatches_count": 400,
    "status_mismatches_count": 250,
    "total_discrepancies": 1000
  }
}
```

### GET /api/discrepancies

Retrieves paginated discrepancy data from the database.

**Parameters**:
- `page`: Page number (default: 1)
- `pageSize`: Number of items per page (default: 50, max: 1000)
- `type`: Filter by discrepancy type (optional): "MISSING_IN_A", "MISSING_IN_B", "AMOUNT_MISMATCH", "STATUS_MISMATCH"

**Response**:
```json
{
  "page": 1,
  "page_size": 50,
  "total_pages": 20,
  "total_items": 1000,
  "discrepancies": [
    {
      "transaction_id": "TXN-00001234",
      "type": "MISSING_IN_A",
      "details": "Transaction TXN-00001234 found in System B but missing in System A"
    },
    // ... more items
  ]
}
```

## Performance Considerations

This system is designed to handle large CSV files efficiently:

1. **Streaming Processing**: Files are processed in a streaming fashion to minimize memory usage
2. **Efficient Data Structures**: Hash maps (Go maps) are used for O(1) lookups
3. **Pagination**: Results are paginated to handle large result sets
4. **Database Indexing**: Key fields are indexed for fast query performance
5. **Optimized Frontend**: The frontend uses virtual pagination to efficiently render large datasets

## Extending the System

### Adding New Discrepancy Types

To add new discrepancy types:

1. Update the backend `reconcileTransactions` function to detect new discrepancies
2. Add a new tab in the frontend interface
3. Update the database schema if necessary

### Supporting Additional File Formats

The system can be extended to support additional file formats:

1. Create a new parser function for the desired format
2. Implement auto-detection of file format or add a format selection option

<!-- ## License

This project is licensed under the MIT License - see the LICENSE file for details. -->