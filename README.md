# dupay

[![CI](https://github.com/rasulov-emirlan/dupay/actions/workflows/ci.yml/badge.svg)](https://github.com/rasulov-emirlan/dupay/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Detect duplicate payments across multiple bank statement PDFs.

## Features

- Extracts transactions from PDF bank statements
- Supports multiple bank formats (extensible)
- Compares transactions across different banks
- Configurable time and amount tolerances
- Reports potential duplicates with detailed information

## Supported Banks

| Bank | Country | Status |
|------|---------|--------|
| Optima Bank | Kyrgyzstan | Supported |
| Mbank | Kyrgyzstan | Supported |

## Installation

### From source

```bash
go install github.com/rasulov-emirlan/dupay@latest
```

### Build manually

```bash
git clone https://github.com/rasulov-emirlan/dupay.git
cd dupay
go build -o dupay .
```

## Usage

```bash
dupay [options] <pdf1> <pdf2> [pdf3...]
```

### Options

| Flag | Description | Default |
|------|-------------|---------|
| `-time` | Maximum time difference between transactions | `1m` |
| `-amount` | Maximum amount difference in KGS | `1.0` |
| `-version` | Print version information | - |

### Examples

Basic usage with two bank statements:
```bash
dupay optima.pdf mbank.pdf
```

With custom tolerances:
```bash
dupay -time 2m -amount 5 optima.pdf mbank.pdf
```

Multiple statements from the same period:
```bash
dupay -time 1m optima_jan.pdf optima_feb.pdf mbank_q1.pdf
```

## How It Works

1. **PDF Parsing**: Extracts text content from each PDF file
2. **Bank Detection**: Automatically identifies the bank format based on content patterns
3. **Transaction Extraction**: Parses transactions using bank-specific parsers
4. **Deduplication**: Removes duplicate entries within the same bank (for overlapping statement periods)
5. **Cross-Bank Comparison**: Compares transactions across different banks looking for:
   - Similar timestamps (within configured tolerance)
   - Similar amounts (within configured tolerance)
   - Same currency
   - Both are debit transactions (outgoing payments)

## Adding Support for New Banks

To add support for a new bank, implement the `BankParser` interface:

```go
type BankParser interface {
    Parse(content string) ([]Transaction, error)
    CanParse(content string) bool
    BankName() string
}
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed instructions.

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
