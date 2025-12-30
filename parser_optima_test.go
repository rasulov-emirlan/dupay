package main

import (
	"testing"
)

func TestOptimaParser_BankName(t *testing.T) {
	parser := NewOptimaParser()
	expected := "Optima Bank"
	if parser.BankName() != expected {
		t.Errorf("expected %q, got %q", expected, parser.BankName())
	}
}

func TestOptimaParser_CanParse(t *testing.T) {
	parser := NewOptimaParser()

	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "contains Optima Bank",
			content:  "Statement from Optima Bank for January",
			expected: true,
		},
		{
			name:     "contains OptimaBank",
			content:  "OptimaBank statement",
			expected: true,
		},
		{
			name:     "contains optimabank.kg",
			content:  "Visit optimabank.kg for more info",
			expected: true,
		},
		{
			name:     "no Optima references",
			content:  "Statement from Another Bank",
			expected: false,
		},
		{
			name:     "empty content",
			content:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.CanParse(tt.content)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestOptimaParser_Parse(t *testing.T) {
	parser := NewOptimaParser()

	tests := []struct {
		name             string
		content          string
		expectedCount    int
		expectedAmount   float64
		expectedCurrency string
	}{
		{
			name:          "empty content",
			content:       "",
			expectedCount: 0,
		},
		{
			name: "single transaction",
			content: `Optima Bank Statement
15.01.2025
10:30
Payment to merchant
-1 500.00
KGS
0
KGS`,
			expectedCount:    1,
			expectedAmount:   -1500.0,
			expectedCurrency: "KGS",
		},
		{
			name: "transaction with credit (positive amount)",
			content: `Optima Bank Statement
15.01.2025
10:30
Incoming transfer
5 000.00
KGS
0
KGS`,
			expectedCount:    1,
			expectedAmount:   5000.0,
			expectedCurrency: "KGS",
		},
		{
			name: "multiple transactions",
			content: `Optima Bank Statement
15.01.2025
10:30
First payment
-1 000.00
KGS
0
KGS
15.01.2025
14:45
Second payment
-2 000.00
KGS
0
KGS`,
			expectedCount: 2,
		},
		{
			name: "skip header lines",
			content: `Optima Bank Statement
Date
Details
of
operations
Operation
amount
15.01.2025
10:30
Real transaction
-500.00
KGS
0
KGS`,
			expectedCount:  1,
			expectedAmount: -500.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transactions, err := parser.Parse(tt.content)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(transactions) != tt.expectedCount {
				t.Errorf("expected %d transactions, got %d", tt.expectedCount, len(transactions))
			}
			if tt.expectedCount > 0 && tt.expectedAmount != 0 {
				if transactions[0].Amount != tt.expectedAmount {
					t.Errorf("expected amount %v, got %v", tt.expectedAmount, transactions[0].Amount)
				}
			}
			if tt.expectedCount > 0 && tt.expectedCurrency != "" {
				if transactions[0].Currency != tt.expectedCurrency {
					t.Errorf("expected currency %v, got %v", tt.expectedCurrency, transactions[0].Currency)
				}
			}
		})
	}
}

func TestNormalizeSpaces(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "regular spaces",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "non-breaking space",
			input:    "hello\u00a0world",
			expected: "hello world",
		},
		{
			name:     "figure space",
			input:    "1\u2007000",
			expected: "1 000",
		},
		{
			name:     "narrow no-break space",
			input:    "1\u202f000",
			expected: "1 000",
		},
		{
			name:     "mixed spaces",
			input:    "1\u00a0000\u2007500\u202f00",
			expected: "1 000 500 00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeSpaces(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestParseOptimaAmount(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name:     "simple amount",
			input:    "100.00",
			expected: 100.0,
		},
		{
			name:     "negative amount",
			input:    "-100.00",
			expected: -100.0,
		},
		{
			name:     "amount with space separator",
			input:    "1 000.00",
			expected: 1000.0,
		},
		{
			name:     "large amount with spaces",
			input:    "1 234 567.89",
			expected: 1234567.89,
		},
		{
			name:     "negative with spaces",
			input:    "-1 965.84",
			expected: -1965.84,
		},
		{
			name:     "amount with non-breaking space",
			input:    "1\u00a0000.00",
			expected: 1000.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseOptimaAmount(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
