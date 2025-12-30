package main

import (
	"testing"
)

func TestMbankParser_BankName(t *testing.T) {
	parser := NewMbankParser()
	expected := "Mbank"
	if parser.BankName() != expected {
		t.Errorf("expected %q, got %q", expected, parser.BankName())
	}
}

func TestMbankParser_CanParse(t *testing.T) {
	parser := NewMbankParser()

	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "contains mbank.kg",
			content:  "Statement from mbank.kg",
			expected: true,
		},
		{
			name:     "contains Mbank",
			content:  "Mbank statement for January",
			expected: true,
		},
		{
			name:     "contains МБАНК (Cyrillic)",
			content:  "Выписка МБАНК за январь",
			expected: true,
		},
		{
			name:     "no Mbank references",
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

func TestMbankParser_Parse(t *testing.T) {
	parser := NewMbankParser()

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
			content: `Mbank Statement
24.12.2025 12:02 Оплата в магазине - 1 018,00`,
			expectedCount:    1,
			expectedAmount:   -1018.0,
			expectedCurrency: "KGS",
		},
		{
			name: "positive transaction",
			content: `Mbank Statement
24.12.2025 14:30 Пополнение счета 5 000,00`,
			expectedCount:    1,
			expectedAmount:   5000.0,
			expectedCurrency: "KGS",
		},
		{
			name: "multiple transactions",
			content: `Mbank Statement
24.12.2025 10:00 First payment - 500,00
24.12.2025 11:00 Second payment - 1 000,00`,
			expectedCount: 2,
		},
		{
			name: "skip header lines",
			content: `Mbank Statement
Выписка по счету
За период с 01.12.2025 по 31.12.2025
Дата операции
Описание операции
Сумма операции
24.12.2025 10:00 Real transaction - 500,00`,
			expectedCount:  1,
			expectedAmount: -500.0,
		},
		{
			name: "skip footer lines",
			content: `Mbank Statement
24.12.2025 10:00 Payment - 500,00
Всего списаний: 500,00
Для проверки подлинности`,
			expectedCount:  1,
			expectedAmount: -500.0,
		},
		{
			name: "multiline description",
			content: `Mbank Statement
24.12.2025 10:00 First line of
description continues here - 500,00`,
			expectedCount: 1,
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

func TestParseMbankAmount(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name:     "simple amount",
			input:    "100,00",
			expected: 100.0,
		},
		{
			name:     "negative amount",
			input:    "-100,00",
			expected: -100.0,
		},
		{
			name:     "negative with space after minus",
			input:    "- 100,00",
			expected: -100.0,
		},
		{
			name:     "amount with space separator",
			input:    "1 000,00",
			expected: 1000.0,
		},
		{
			name:     "large amount with spaces",
			input:    "1 234 567,89",
			expected: 1234567.89,
		},
		{
			name:     "negative with spaces",
			input:    "- 1 018,00",
			expected: -1018.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseMbankAmount(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMbankParser_TransactionDetails(t *testing.T) {
	parser := NewMbankParser()

	content := `Mbank Statement
24.12.2025 12:02 Оплата покупки в магазине - 1 500,00`

	transactions, err := parser.Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(transactions) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(transactions))
	}

	tx := transactions[0]

	// Check bank name
	if tx.Bank != "Mbank" {
		t.Errorf("expected bank 'Mbank', got %q", tx.Bank)
	}

	// Check currency
	if tx.Currency != "KGS" {
		t.Errorf("expected currency 'KGS', got %q", tx.Currency)
	}

	// Check description contains expected text
	if tx.Description == "" {
		t.Error("expected non-empty description")
	}

	// Check datetime
	if tx.DateTime.Day() != 24 || tx.DateTime.Month() != 12 || tx.DateTime.Year() != 2025 {
		t.Errorf("unexpected date: %v", tx.DateTime)
	}
	if tx.DateTime.Hour() != 12 || tx.DateTime.Minute() != 2 {
		t.Errorf("unexpected time: %v", tx.DateTime)
	}
}
