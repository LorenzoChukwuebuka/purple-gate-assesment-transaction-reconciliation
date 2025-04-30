package main

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"time"
)

func main() {
	// Set a seed for random number generation
	rand.Seed(time.Now().UnixNano())

	// Define parameters
	numRecords := 1000    // Change this to generate more or fewer records
	discrepancyRate := 0.05 // 5% of records will have discrepancies

	// Generate transaction IDs
	transactionIDs := make([]string, numRecords)
	for i := 0; i < numRecords; i++ {
		transactionIDs[i] = fmt.Sprintf("TXN-%08d", i+1)
	}

	// Create file A
	if err := generateFileA(transactionIDs, discrepancyRate); err != nil {
		fmt.Printf("Error generating file A: %v\n", err)
		return
	}

	// Create file B with some discrepancies
	if err := generateFileB(transactionIDs, discrepancyRate); err != nil {
		fmt.Printf("Error generating file B: %v\n", err)
		return
	}

	fmt.Println("Sample data generated successfully!")
	fmt.Printf("Generated %d transactions with approximately %.1f%% discrepancies\n",
		numRecords, discrepancyRate*100)
}

func generateFileA(transactionIDs []string, discrepancyRate float64) error {
	file, err := os.Create("source_system_a.csv")
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	err = writer.Write([]string{
		"transaction_id",
		"timestamp",
		"amount",
		"currency",
		"status",
	})
	if err != nil {
		return err
	}

	// Skip some records to create "missing in A" discrepancies
	skipIndices := make(map[int]bool)
	for i := 0; i < int(float64(len(transactionIDs))*discrepancyRate*0.2); i++ {
		skipIndex := rand.Intn(len(transactionIDs))
		skipIndices[skipIndex] = true
	}

	// Write data rows
	for i, txID := range transactionIDs {
		// Skip some transactions
		if skipIndices[i] {
			continue
		}

		timestamp := time.Now().Add(-time.Duration(rand.Intn(24*30)) * time.Hour).Format(time.RFC3339)
		amount := 100 + rand.Float64()*9900
		currency := "USD"
		if rand.Float32() < 0.1 {
			currency = "EUR"
		} else if rand.Float32() < 0.05 {
			currency = "GBP"
		}

		status := "SUCCESS"
		if rand.Float32() < 0.08 {
			status = "FAILED"
		} else if rand.Float32() < 0.03 {
			status = "PENDING"
		}

		err = writer.Write([]string{
			txID,
			timestamp,
			fmt.Sprintf("%.2f", amount),
			currency,
			status,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func generateFileB(transactionIDs []string, discrepancyRate float64) error {
	file, err := os.Create("source_system_b.csv")
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	err = writer.Write([]string{
		"transaction_id",
		"timestamp",
		"amount",
		"currency",
		"status",
	})
	if err != nil {
		return err
	}

	// Skip some records to create "missing in B" discrepancies
	skipIndices := make(map[int]bool)
	for i := 0; i < int(float64(len(transactionIDs))*discrepancyRate*0.2); i++ {
		skipIndex := rand.Intn(len(transactionIDs))
		skipIndices[skipIndex] = true
	}

	// Prepare indices for amount and status discrepancies
	amountDiscrepancyIndices := make(map[int]bool)
	for i := 0; i < int(float64(len(transactionIDs))*discrepancyRate*0.4); i++ {
		index := rand.Intn(len(transactionIDs))
		amountDiscrepancyIndices[index] = true
	}

	statusDiscrepancyIndices := make(map[int]bool)
	for i := 0; i < int(float64(len(transactionIDs))*discrepancyRate*0.2); i++ {
		index := rand.Intn(len(transactionIDs))
		statusDiscrepancyIndices[index] = true
	}

	// Write data rows
	for i, txID := range transactionIDs {
		// Skip some transactions
		if skipIndices[i] {
			continue
		}

		timestamp := time.Now().Add(-time.Duration(rand.Intn(24*30)) * time.Hour).Format(time.RFC3339)

		// Base amount
		amount := 100 + rand.Float64()*9900

		// Create some amount discrepancies
		if amountDiscrepancyIndices[i] {
			// Modify the amount slightly to create a discrepancy
			discrepancyFactor := 1.0 + (rand.Float64()*0.1 - 0.05) // ±5%
			amount *= discrepancyFactor
		}

		currency := "USD"
		if rand.Float32() < 0.1 {
			currency = "EUR"
		} else if rand.Float32() < 0.05 {
			currency = "GBP"
		}

		status := "SUCCESS"
		if rand.Float32() < 0.08 {
			status = "FAILED"
		} else if rand.Float32() < 0.03 {
			status = "PENDING"
		}

		// Create some status discrepancies
		if statusDiscrepancyIndices[i] {
			// Change the status to create a discrepancy
			statuses := []string{"SUCCESS", "FAILED", "PENDING"}
			newStatusIndex := rand.Intn(len(statuses))
			if statuses[newStatusIndex] == status {
				newStatusIndex = (newStatusIndex + 1) % len(statuses)
			}
			status = statuses[newStatusIndex]
		}

		err = writer.Write([]string{
			txID,
			timestamp,
			fmt.Sprintf("%.2f", amount),
			currency,
			status,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
