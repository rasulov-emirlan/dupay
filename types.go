package main

import (
	"time"
)

// Transaction represents a single bank transaction extracted from a PDF statement.
type Transaction struct {
	// DateTime is the date and time when the transaction occurred.
	DateTime time.Time
	// Description contains the transaction details/memo from the bank statement.
	Description string
	// Amount is the transaction value. Negative for debits (outgoing), positive for credits (incoming).
	Amount float64
	// Currency is the ISO currency code (e.g., "KGS", "USD").
	Currency string
	// Bank is the name of the bank this transaction came from.
	Bank string
	// RawLine contains the original text from the PDF for debugging purposes.
	RawLine string
}

// BankParser defines the interface for parsing bank-specific PDF statement formats.
// Implement this interface to add support for a new bank.
type BankParser interface {
	// Parse extracts transactions from the PDF text content.
	// Returns a slice of transactions and any error encountered during parsing.
	Parse(content string) ([]Transaction, error)
	// BankName returns the human-readable name of the bank.
	BankName() string
	// CanParse checks if this parser can handle the given PDF content.
	// It should return true if the content matches the expected bank format.
	CanParse(content string) bool
}

// DuplicateMatch represents a potential duplicate payment found across different banks.
type DuplicateMatch struct {
	// Transaction1 is the first transaction in the potential duplicate pair.
	Transaction1 Transaction
	// Transaction2 is the second transaction in the potential duplicate pair.
	Transaction2 Transaction
	// TimeDiff is the absolute time difference between the two transactions.
	TimeDiff time.Duration
	// AmountDiff is the absolute difference in amounts between the two transactions.
	AmountDiff float64
}
