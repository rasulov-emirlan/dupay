package main

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

type OptimaParser struct{}

func NewOptimaParser() *OptimaParser {
	return &OptimaParser{}
}

func (p *OptimaParser) BankName() string {
	return "Optima Bank"
}

func (p *OptimaParser) CanParse(content string) bool {
	return strings.Contains(content, "Optima Bank") ||
		strings.Contains(content, "OptimaBank") ||
		strings.Contains(content, "optimabank.kg")
}

// normalizeSpaces replaces non-breaking spaces and other whitespace with regular spaces
func normalizeSpaces(s string) string {
	// Replace non-breaking space (U+00A0) with regular space
	s = strings.ReplaceAll(s, "\u00a0", " ")
	// Replace other unicode spaces
	s = strings.ReplaceAll(s, "\u2007", " ") // figure space
	s = strings.ReplaceAll(s, "\u202f", " ") // narrow no-break space
	return s
}

func (p *OptimaParser) Parse(content string) ([]Transaction, error) {
	var transactions []Transaction

	// Normalize spaces in content
	content = normalizeSpaces(content)

	lines := strings.Split(content, "\n")

	// Regex patterns
	datePattern := regexp.MustCompile(`^(\d{2}\.\d{2}\.\d{4})\s*$`)
	timePattern := regexp.MustCompile(`^(\d{2}:\d{2})$`)
	// Amount can be negative like "-1 965.84" or positive like "318 273.38"
	amountPattern := regexp.MustCompile(`^(-?[\d\s]+(?:\.\d+)?)\s*$`)

	i := 0
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])

		// Look for date pattern
		if !datePattern.MatchString(line) {
			i++
			continue
		}

		currentDate := strings.TrimSpace(line)
		i++

		// Next should be time
		if i >= len(lines) {
			break
		}
		timeLine := strings.TrimSpace(lines[i])
		if !timePattern.MatchString(timeLine) {
			continue
		}
		currentTime := timeLine
		i++

		// Collect description lines until we hit an amount
		var descLines []string
		var amountKGS float64
		foundAmount := false

		for i < len(lines) {
			line := strings.TrimSpace(lines[i])

			// Skip empty lines
			if line == "" {
				i++
				continue
			}

			// Check if this looks like an amount
			if amountPattern.MatchString(line) {
				// Peek ahead to see if next non-empty line is "KGS" or "USD"
				j := i + 1
				for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
					j++
				}
				if j < len(lines) {
					nextLine := strings.TrimSpace(lines[j])
					if nextLine == "KGS" || nextLine == "USD" {
						// This is an amount
						amt := parseOptimaAmount(line)

						if nextLine == "USD" {
							// Skip USD amount, look for KGS amount after
							i = j + 1
							// Find the KGS amount
							for i < len(lines) {
								l := strings.TrimSpace(lines[i])
								if l == "" {
									i++
									continue
								}
								if amountPattern.MatchString(l) {
									k := i + 1
									for k < len(lines) && strings.TrimSpace(lines[k]) == "" {
										k++
									}
									if k < len(lines) && strings.TrimSpace(lines[k]) == "KGS" {
										amountKGS = parseOptimaAmount(l)
										i = k + 1
										foundAmount = true
										break
									}
								}
								i++
							}
							break
						} else {
							// This is the KGS amount
							amountKGS = amt
							i = j + 1
							foundAmount = true
							break
						}
					}
				}
			}

			// Skip page numbers and headers
			if strings.Contains(line, "/6") ||
				line == "Date" || line == "Details" || line == "of" ||
				line == "operations" || line == "Operation" || line == "amount" ||
				line == "Fee" || line == "0" {
				i++
				continue
			}

			// This is part of description
			descLines = append(descLines, line)
			i++
		}

		if !foundAmount {
			continue
		}

		// Skip fee (0 KGS)
		for i < len(lines) {
			line := strings.TrimSpace(lines[i])
			if line == "0" || line == "KGS" || line == "" {
				i++
			} else {
				break
			}
		}

		// Parse datetime
		dateTime, err := time.Parse("02.01.2006 15:04", currentDate+" "+currentTime)
		if err != nil {
			continue
		}

		description := strings.Join(descLines, " ")

		transactions = append(transactions, Transaction{
			DateTime:    dateTime,
			Description: description,
			Amount:      amountKGS,
			Currency:    "KGS",
			Bank:        p.BankName(),
			RawLine:     description,
		})
	}

	return transactions, nil
}

func parseOptimaAmount(s string) float64 {
	// Normalize spaces first
	s = normalizeSpaces(s)
	// Remove all spaces (thousands separator)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.TrimSpace(s)

	amount, _ := strconv.ParseFloat(s, 64)
	return amount
}
