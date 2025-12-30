package main

import (
	"fmt"
	"math"
	"time"
)

// deduplicateTransactions removes duplicate transactions from the same bank
// (same bank, same datetime, same amount - likely from overlapping statement periods)
func deduplicateTransactions(transactions []Transaction) []Transaction {
	seen := make(map[string]bool)
	var result []Transaction

	for _, t := range transactions {
		// Create a unique key for each transaction
		key := fmt.Sprintf("%s|%s|%.2f", t.Bank, t.DateTime.Format("2006-01-02 15:04"), t.Amount)
		if !seen[key] {
			seen[key] = true
			result = append(result, t)
		}
	}

	return result
}

// FindDuplicates finds potential duplicate transactions across different banks
// Parameters:
//   - transactions: all transactions from all banks
//   - maxTimeDiff: maximum time difference to consider (e.g., 1 minute)
//   - maxAmountDiff: maximum amount difference in KGS (e.g., 1.0)
func FindDuplicates(transactions []Transaction, maxTimeDiff time.Duration, maxAmountDiff float64) []DuplicateMatch {
	// First, deduplicate transactions from overlapping statement periods
	transactions = deduplicateTransactions(transactions)

	var matches []DuplicateMatch

	// Only compare transactions from different banks
	for i := 0; i < len(transactions); i++ {
		for j := i + 1; j < len(transactions); j++ {
			t1 := transactions[i]
			t2 := transactions[j]

			// Skip if same bank
			if t1.Bank == t2.Bank {
				continue
			}

			// Skip if different currencies
			if t1.Currency != t2.Currency {
				continue
			}

			// Both should be debits (negative amounts) for duplicate payment detection
			if t1.Amount >= 0 || t2.Amount >= 0 {
				continue
			}

			// Check time difference
			timeDiff := t1.DateTime.Sub(t2.DateTime)
			if timeDiff < 0 {
				timeDiff = -timeDiff
			}
			if timeDiff > maxTimeDiff {
				continue
			}

			// Check amount difference (compare absolute values since both are negative)
			amountDiff := math.Abs(math.Abs(t1.Amount) - math.Abs(t2.Amount))
			if amountDiff > maxAmountDiff {
				continue
			}

			matches = append(matches, DuplicateMatch{
				Transaction1: t1,
				Transaction2: t2,
				TimeDiff:     timeDiff,
				AmountDiff:   amountDiff,
			})
		}
	}

	return matches
}
