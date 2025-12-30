# Contributing to dupay

Thank you for your interest in contributing to dupay! This document provides guidelines and information for contributors.

## How to Contribute

### Reporting Bugs

If you find a bug, please open an issue on GitHub with:
- A clear, descriptive title
- Steps to reproduce the problem
- Expected behavior vs actual behavior
- Your environment (Go version, OS)

### Suggesting Features

Feature suggestions are welcome! Please open an issue with:
- A clear description of the feature
- Use case and motivation
- Any implementation ideas you have

### Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`make test`)
5. Run linter (`make lint`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to your branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

## Adding Support for New Banks

One of the most valuable contributions is adding support for new banks. Here's how:

### 1. Create a new parser file

Create `parser_<bankname>.go`:

```go
package main

import (
    "regexp"
    "strings"
    "time"
)

type BanknameParser struct{}

func NewBanknameParser() *BanknameParser {
    return &BanknameParser{}
}

func (p *BanknameParser) BankName() string {
    return "Bank Name"
}

func (p *BanknameParser) CanParse(content string) bool {
    // Return true if this parser can handle the content
    return strings.Contains(content, "unique-bank-identifier")
}

func (p *BanknameParser) Parse(content string) ([]Transaction, error) {
    var transactions []Transaction

    // Parse the PDF content and extract transactions
    // Each transaction should have:
    // - DateTime: time.Time
    // - Description: string
    // - Amount: float64 (negative for debits, positive for credits)
    // - Currency: string (e.g., "KGS", "USD")
    // - Bank: string (use p.BankName())

    return transactions, nil
}
```

### 2. Register the parser

Add your parser to the list in `main.go`:

```go
parsers := []BankParser{
    NewOptimaParser(),
    NewMbankParser(),
    NewBanknameParser(), // Add your parser here
}
```

### 3. Write tests

Create `parser_<bankname>_test.go` with tests for:
- `BankName()` returns correct name
- `CanParse()` correctly identifies bank's PDFs
- `Parse()` extracts transactions correctly

### 4. Update documentation

- Add the bank to the supported banks table in README.md
- Include any special notes about the bank's format

## Code Style

- Follow standard Go formatting (`go fmt`)
- Use meaningful variable and function names
- Add comments for complex logic
- Keep functions focused and small
- Handle errors appropriately

## Testing

- Write tests for new functionality
- Ensure all tests pass before submitting PR
- Aim for good test coverage on new code

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage
```

## Commit Messages

- Use clear, descriptive commit messages
- Start with a verb (Add, Fix, Update, Remove, etc.)
- Keep the first line under 72 characters

Examples:
- `Add support for BankX PDF statements`
- `Fix amount parsing for transactions with spaces`
- `Update README with new installation instructions`

## Code of Conduct

- Be respectful and inclusive
- Provide constructive feedback
- Focus on the code, not the person

## Questions?

Feel free to open an issue if you have questions about contributing.

Thank you for contributing!
