package main

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

type MbankParser struct{}

func NewMbankParser() *MbankParser {
	return &MbankParser{}
}

func (p *MbankParser) BankName() string {
	return "Mbank"
}

func (p *MbankParser) CanParse(content string) bool {
	return strings.Contains(content, "mbank.kg") ||
		strings.Contains(content, "Mbank") ||
		strings.Contains(content, "МБАНК")
}

func (p *MbankParser) Parse(content string) ([]Transaction, error) {
	var transactions []Transaction

	lines := strings.Split(content, "\n")

	// Regex patterns
	// Match date-time at start of line like "24.12.2025 12:02"
	dateTimePattern := regexp.MustCompile(`^(\d{2}\.\d{2}\.\d{4})\s+(\d{2}:\d{2})`)
	// Match amount like "- 1 018,00" or "-1018,00" or "1 018,00"
	amountPattern := regexp.MustCompile(`(-?\s*[\d\s]+,\d{2})$`)

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// Skip headers and footers
		if line == "" ||
			strings.HasPrefix(line, "Выписка по счету") ||
			strings.HasPrefix(line, "За период") ||
			strings.HasPrefix(line, "Дата формирования") ||
			strings.HasPrefix(line, "Клиент") ||
			strings.HasPrefix(line, "Баланс") ||
			line == "Дата операции" ||
			line == "Описание операции" ||
			line == "Сумма операции" ||
			strings.HasPrefix(line, "Всего списаний") ||
			strings.HasPrefix(line, "Всего пополнений") ||
			strings.HasPrefix(line, "Для проверки") ||
			strings.HasPrefix(line, "Данная информация") ||
			strings.HasPrefix(line, "С0082") ||
			strings.HasPrefix(line, "Телефон") ||
			strings.HasPrefix(line, "Факс") ||
			strings.HasPrefix(line, "E-mail") ||
			strings.HasPrefix(line, "www.") ||
			strings.Contains(line, "РАСУЛОВ") ||
			strings.Contains(line, "KGS KGS") {
			continue
		}

		// Check if line starts with date-time
		if dateTimePattern.MatchString(line) {
			matches := dateTimePattern.FindStringSubmatch(line)
			if len(matches) < 3 {
				continue
			}

			dateStr := matches[1]
			timeStr := matches[2]

			// Get the description (rest of the line after datetime)
			restOfLine := strings.TrimSpace(line[len(matches[0]):])

			// Collect full description and amount
			// Description might continue on next lines, amount is at the end
			fullText := restOfLine

			// Look ahead for continuation lines (lines that don't start with date)
			for j := i + 1; j < len(lines); j++ {
				nextLine := strings.TrimSpace(lines[j])
				if nextLine == "" {
					continue
				}
				// If next line starts with a date, stop
				if dateTimePattern.MatchString(nextLine) {
					break
				}
				// Skip obvious header/footer lines
				if strings.HasPrefix(nextLine, "Всего") ||
					strings.HasPrefix(nextLine, "Для проверки") ||
					strings.HasPrefix(nextLine, "Данная информация") {
					break
				}
				fullText += " " + nextLine
				i = j // Skip these lines in main loop
			}

			// Try to extract amount from the end
			amountMatches := amountPattern.FindStringSubmatch(fullText)
			if len(amountMatches) < 2 {
				continue
			}

			amountStr := amountMatches[1]
			amount := parseMbankAmount(amountStr)

			// Skip if amount is 0
			if amount == 0 {
				continue
			}

			// Description is everything before the amount
			amountIdx := strings.LastIndex(fullText, amountMatches[0])
			description := strings.TrimSpace(fullText[:amountIdx])

			// Parse datetime
			dateTime, err := time.Parse("02.01.2006 15:04", dateStr+" "+timeStr)
			if err != nil {
				continue
			}

			transactions = append(transactions, Transaction{
				DateTime:    dateTime,
				Description: description,
				Amount:      amount,
				Currency:    "KGS",
				Bank:        p.BankName(),
				RawLine:     line,
			})
		}
	}

	return transactions, nil
}

func parseMbankAmount(s string) float64 {
	// Remove spaces
	s = strings.ReplaceAll(s, " ", "")
	// Replace comma with dot for decimal
	s = strings.ReplaceAll(s, ",", ".")
	s = strings.TrimSpace(s)

	// Handle "- 1018.00" format (negative with space after minus)
	s = strings.ReplaceAll(s, "- ", "-")

	amount, _ := strconv.ParseFloat(s, 64)
	return amount
}
