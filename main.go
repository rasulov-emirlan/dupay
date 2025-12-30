package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ledongthuc/pdf"
)

// Version is set at build time via -ldflags
var Version = "dev"

func main() {
	// CLI flags
	maxTimeDiff := flag.Duration("time", time.Minute, "Maximum time difference between transactions (e.g., 1m, 2m)")
	maxAmountDiff := flag.Float64("amount", 1.0, "Maximum amount difference in KGS")
	showVersion := flag.Bool("version", false, "Print version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("dupay version %s\n", Version)
		os.Exit(0)
	}

	// Get PDF files from arguments
	pdfFiles := flag.Args()
	if len(pdfFiles) < 2 {
		fmt.Println("Usage: dupay [options] <pdf1> <pdf2> [pdf3...]")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nExample:")
		fmt.Println("  dupay -time 1m -amount 1 optima.pdf mbank.pdf")
		os.Exit(1)
	}

	// Register all available parsers
	parsers := []BankParser{
		NewOptimaParser(),
		NewMbankParser(),
	}

	// Parse all PDFs
	var allTransactions []Transaction

	for _, pdfFile := range pdfFiles {
		fmt.Printf("Processing: %s\n", filepath.Base(pdfFile))

		// Extract text from PDF
		content, err := extractPDFText(pdfFile)
		if err != nil {
			fmt.Printf("  Error reading PDF: %v\n", err)
			continue
		}

		// Find matching parser
		var matchedParser BankParser
		for _, parser := range parsers {
			if parser.CanParse(content) {
				matchedParser = parser
				break
			}
		}

		if matchedParser == nil {
			fmt.Printf("  Warning: No parser found for this PDF format\n")
			continue
		}

		fmt.Printf("  Detected: %s\n", matchedParser.BankName())

		// Parse transactions
		transactions, err := matchedParser.Parse(content)
		if err != nil {
			fmt.Printf("  Error parsing: %v\n", err)
			continue
		}

		fmt.Printf("  Found %d transactions\n", len(transactions))
		allTransactions = append(allTransactions, transactions...)
	}

	fmt.Printf("\nTotal transactions: %d\n", len(allTransactions))
	fmt.Printf("Looking for duplicates (time diff <= %v, amount diff <= %.2f KGS)...\n\n", *maxTimeDiff, *maxAmountDiff)

	// Find duplicates
	duplicates := FindDuplicates(allTransactions, *maxTimeDiff, *maxAmountDiff)

	if len(duplicates) == 0 {
		fmt.Println("No potential duplicates found.")
		return
	}

	fmt.Printf("Found %d potential duplicate(s):\n\n", len(duplicates))

	for i, dup := range duplicates {
		fmt.Printf("=== Duplicate #%d ===\n", i+1)
		fmt.Printf("Time difference: %v\n", dup.TimeDiff)
		fmt.Printf("Amount difference: %.2f KGS\n\n", dup.AmountDiff)

		fmt.Printf("Transaction 1 (%s):\n", dup.Transaction1.Bank)
		fmt.Printf("  Date/Time: %s\n", dup.Transaction1.DateTime.Format("02.01.2006 15:04"))
		fmt.Printf("  Amount: %.2f %s\n", dup.Transaction1.Amount, dup.Transaction1.Currency)
		fmt.Printf("  Description: %s\n\n", truncateString(dup.Transaction1.Description, 80))

		fmt.Printf("Transaction 2 (%s):\n", dup.Transaction2.Bank)
		fmt.Printf("  Date/Time: %s\n", dup.Transaction2.DateTime.Format("02.01.2006 15:04"))
		fmt.Printf("  Amount: %.2f %s\n", dup.Transaction2.Amount, dup.Transaction2.Currency)
		fmt.Printf("  Description: %s\n", truncateString(dup.Transaction2.Description, 80))
		fmt.Println(strings.Repeat("-", 60))
	}

	// Summary
	var totalDuplicateAmount float64
	for _, dup := range duplicates {
		// Use the average of both amounts
		totalDuplicateAmount += (dup.Transaction1.Amount + dup.Transaction2.Amount) / 2
	}
	fmt.Printf("\nTotal potential duplicate amount: %.2f KGS\n", totalDuplicateAmount)
}

// extractPDFText extracts all text content from a PDF file
func extractPDFText(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var buf strings.Builder
	for i := 1; i <= r.NumPage(); i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			continue
		}
		text, err := p.GetPlainText(nil)
		if err != nil {
			continue
		}
		buf.WriteString(text)
		buf.WriteString("\n")
	}

	return buf.String(), nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
