package main

import (
	"testing"
	"time"
)

func TestDeduplicateTransactions(t *testing.T) {
	baseTime := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		input    []Transaction
		expected int
	}{
		{
			name:     "empty input",
			input:    []Transaction{},
			expected: 0,
		},
		{
			name: "no duplicates",
			input: []Transaction{
				{Bank: "BankA", DateTime: baseTime, Amount: -100.0},
				{Bank: "BankA", DateTime: baseTime.Add(time.Hour), Amount: -200.0},
			},
			expected: 2,
		},
		{
			name: "exact duplicates same bank",
			input: []Transaction{
				{Bank: "BankA", DateTime: baseTime, Amount: -100.0},
				{Bank: "BankA", DateTime: baseTime, Amount: -100.0},
			},
			expected: 1,
		},
		{
			name: "same time different banks - not duplicates for dedup",
			input: []Transaction{
				{Bank: "BankA", DateTime: baseTime, Amount: -100.0},
				{Bank: "BankB", DateTime: baseTime, Amount: -100.0},
			},
			expected: 2,
		},
		{
			name: "multiple duplicates",
			input: []Transaction{
				{Bank: "BankA", DateTime: baseTime, Amount: -100.0},
				{Bank: "BankA", DateTime: baseTime, Amount: -100.0},
				{Bank: "BankA", DateTime: baseTime, Amount: -100.0},
				{Bank: "BankA", DateTime: baseTime.Add(time.Hour), Amount: -200.0},
				{Bank: "BankA", DateTime: baseTime.Add(time.Hour), Amount: -200.0},
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deduplicateTransactions(tt.input)
			if len(result) != tt.expected {
				t.Errorf("expected %d transactions, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestFindDuplicates(t *testing.T) {
	baseTime := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name          string
		transactions  []Transaction
		maxTimeDiff   time.Duration
		maxAmountDiff float64
		expected      int
	}{
		{
			name:          "empty input",
			transactions:  []Transaction{},
			maxTimeDiff:   time.Minute,
			maxAmountDiff: 1.0,
			expected:      0,
		},
		{
			name: "single transaction",
			transactions: []Transaction{
				{Bank: "BankA", DateTime: baseTime, Amount: -100.0, Currency: "KGS"},
			},
			maxTimeDiff:   time.Minute,
			maxAmountDiff: 1.0,
			expected:      0,
		},
		{
			name: "same bank - no match",
			transactions: []Transaction{
				{Bank: "BankA", DateTime: baseTime, Amount: -100.0, Currency: "KGS"},
				{Bank: "BankA", DateTime: baseTime, Amount: -100.0, Currency: "KGS"},
			},
			maxTimeDiff:   time.Minute,
			maxAmountDiff: 1.0,
			expected:      0,
		},
		{
			name: "different banks exact match",
			transactions: []Transaction{
				{Bank: "BankA", DateTime: baseTime, Amount: -100.0, Currency: "KGS"},
				{Bank: "BankB", DateTime: baseTime, Amount: -100.0, Currency: "KGS"},
			},
			maxTimeDiff:   time.Minute,
			maxAmountDiff: 1.0,
			expected:      1,
		},
		{
			name: "different banks within time tolerance",
			transactions: []Transaction{
				{Bank: "BankA", DateTime: baseTime, Amount: -100.0, Currency: "KGS"},
				{Bank: "BankB", DateTime: baseTime.Add(30 * time.Second), Amount: -100.0, Currency: "KGS"},
			},
			maxTimeDiff:   time.Minute,
			maxAmountDiff: 1.0,
			expected:      1,
		},
		{
			name: "different banks outside time tolerance",
			transactions: []Transaction{
				{Bank: "BankA", DateTime: baseTime, Amount: -100.0, Currency: "KGS"},
				{Bank: "BankB", DateTime: baseTime.Add(2 * time.Minute), Amount: -100.0, Currency: "KGS"},
			},
			maxTimeDiff:   time.Minute,
			maxAmountDiff: 1.0,
			expected:      0,
		},
		{
			name: "different banks within amount tolerance",
			transactions: []Transaction{
				{Bank: "BankA", DateTime: baseTime, Amount: -100.0, Currency: "KGS"},
				{Bank: "BankB", DateTime: baseTime, Amount: -100.5, Currency: "KGS"},
			},
			maxTimeDiff:   time.Minute,
			maxAmountDiff: 1.0,
			expected:      1,
		},
		{
			name: "different banks outside amount tolerance",
			transactions: []Transaction{
				{Bank: "BankA", DateTime: baseTime, Amount: -100.0, Currency: "KGS"},
				{Bank: "BankB", DateTime: baseTime, Amount: -105.0, Currency: "KGS"},
			},
			maxTimeDiff:   time.Minute,
			maxAmountDiff: 1.0,
			expected:      0,
		},
		{
			name: "different currencies - no match",
			transactions: []Transaction{
				{Bank: "BankA", DateTime: baseTime, Amount: -100.0, Currency: "KGS"},
				{Bank: "BankB", DateTime: baseTime, Amount: -100.0, Currency: "USD"},
			},
			maxTimeDiff:   time.Minute,
			maxAmountDiff: 1.0,
			expected:      0,
		},
		{
			name: "credit transactions - no match",
			transactions: []Transaction{
				{Bank: "BankA", DateTime: baseTime, Amount: 100.0, Currency: "KGS"},
				{Bank: "BankB", DateTime: baseTime, Amount: 100.0, Currency: "KGS"},
			},
			maxTimeDiff:   time.Minute,
			maxAmountDiff: 1.0,
			expected:      0,
		},
		{
			name: "mixed credit and debit - no match",
			transactions: []Transaction{
				{Bank: "BankA", DateTime: baseTime, Amount: -100.0, Currency: "KGS"},
				{Bank: "BankB", DateTime: baseTime, Amount: 100.0, Currency: "KGS"},
			},
			maxTimeDiff:   time.Minute,
			maxAmountDiff: 1.0,
			expected:      0,
		},
		{
			name: "multiple potential duplicates",
			transactions: []Transaction{
				{Bank: "BankA", DateTime: baseTime, Amount: -100.0, Currency: "KGS"},
				{Bank: "BankB", DateTime: baseTime, Amount: -100.0, Currency: "KGS"},
				{Bank: "BankA", DateTime: baseTime.Add(time.Hour), Amount: -200.0, Currency: "KGS"},
				{Bank: "BankB", DateTime: baseTime.Add(time.Hour), Amount: -200.0, Currency: "KGS"},
			},
			maxTimeDiff:   time.Minute,
			maxAmountDiff: 1.0,
			expected:      2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindDuplicates(tt.transactions, tt.maxTimeDiff, tt.maxAmountDiff)
			if len(result) != tt.expected {
				t.Errorf("expected %d duplicates, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestFindDuplicatesMatchDetails(t *testing.T) {
	baseTime := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)

	transactions := []Transaction{
		{Bank: "Optima", DateTime: baseTime, Amount: -100.0, Currency: "KGS", Description: "Payment 1"},
		{Bank: "Mbank", DateTime: baseTime.Add(30 * time.Second), Amount: -100.5, Currency: "KGS", Description: "Payment 1"},
	}

	result := FindDuplicates(transactions, time.Minute, 1.0)

	if len(result) != 1 {
		t.Fatalf("expected 1 duplicate, got %d", len(result))
	}

	match := result[0]

	// Verify time difference
	expectedTimeDiff := 30 * time.Second
	if match.TimeDiff != expectedTimeDiff {
		t.Errorf("expected time diff %v, got %v", expectedTimeDiff, match.TimeDiff)
	}

	// Verify amount difference
	expectedAmountDiff := 0.5
	if match.AmountDiff != expectedAmountDiff {
		t.Errorf("expected amount diff %v, got %v", expectedAmountDiff, match.AmountDiff)
	}

	// Verify banks are different
	if match.Transaction1.Bank == match.Transaction2.Bank {
		t.Error("matched transactions should be from different banks")
	}
}
