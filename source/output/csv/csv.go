package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/benjohns1/invest-source/app"
	"github.com/benjohns1/invest-source/utils/filesystem"
)

var (
	// CreateFile for creating a local file.
	CreateFile = os.Create

	// NewWriter creates a new CSV writer.
	NewWriter = func(w io.Writer) Writer { return csv.NewWriter(w) }

	// Now implementation.
	Now = time.Now
)

// Writer CSV writing interface.
type Writer interface {
	WriteAll([][]string) error
}

// CSV output implementation.
type Output struct {
	Filename  string
	HeaderRow []string
	Symbols   map[string]struct{}
	Filter    func(app.Quote) bool
	MapRow    func(app.Quote) ([]string, error)
}

// NewGnuCashCSV outputs a CSV formatted for a GnuCash price import.
func NewGnuCashCSV(dir string, symbols []string) (Output, error) {
	if err := filesystem.Mkdir(dir); err != nil {
		return Output{}, err
	}
	symbolMap := make(map[string]struct{}, len(symbols))
	for _, symbol := range symbols {
		symbolMap[symbol] = struct{}{}
	}
	return Output{
		Filename:  fmt.Sprintf("%s/%s.csv", dir, Now().UTC().Format("2006-01-02")),
		HeaderRow: []string{"Namespace", "Symbol", "Date", "Price", "Currency"},
		Filter: func(q app.Quote) bool {
			if _, ok := symbolMap[q.Symbol]; !ok {
				return false
			}
			return true
		},
		MapRow: func(q app.Quote) ([]string, error) {
			return []string{"AMEX", q.Symbol, q.Time.Format("2006-01-02"), q.USD.String(), "USD"}, nil
		},
		Symbols: symbolMap,
	}, nil
}

// Write outputs a list of quotes in CSV format.
func (o Output) Write(quotes []app.Quote) (missing []string, err error) {
	rows, startIdx := o.prepare(len(quotes))

	missing, err = o.bufferRows(quotes, rows, startIdx, missing)
	if err != nil {
		return missing, err
	}

	if err := o.writeBuffer(err, rows); err != nil {
		return missing, err
	}

	return missing, nil
}

func (o Output) prepare(size int) ([][]string, int) {
	if o.HeaderRow != nil {
		size += 1
	}
	rows := make([][]string, size)
	idx := 0
	if o.HeaderRow != nil {
		rows[0] = o.HeaderRow
		idx = 1
	}
	return rows, idx
}

func (o Output) bufferRows(quotes []app.Quote, rows [][]string, idx int, missing []string) ([]string, error) {
	found := make(map[string]struct{}, len(o.Symbols))
	for qNum, q := range quotes {
		if !o.Filter(q) {
			continue
		}
		found[q.Symbol] = struct{}{}
		row, err := o.MapRow(q)
		if err != nil {
			return nil, fmt.Errorf("quote number %d error: %v", qNum, err)
		}
		rows[idx] = row
		idx += 1
	}

	for symbol := range o.Symbols {
		if _, ok := found[symbol]; !ok {
			missing = append(missing, symbol)
		}
	}
	return missing, nil
}

func (o Output) writeBuffer(err error, rows [][]string) error {
	f, err := CreateFile(o.Filename)
	if err != nil {
		return err
	}
	w := NewWriter(f)
	if err := w.WriteAll(rows); err != nil {
		return fmt.Errorf("error writing out CSV file '%s': %v", o.Filename, err)
	}
	return nil
}
