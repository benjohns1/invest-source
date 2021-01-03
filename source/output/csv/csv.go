package csv

import (
	"fmt"
	"time"

	"github.com/benjohns1/invest-source/app"
)

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
	if err := Mkdir(dir); err != nil {
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

// LastRun returns the last run time of writing output.
func (o Output) LastRun() time.Time {
	// TODO: retrieve this from a file, write to it whenever runs
	return time.Time{}
}

// WriteSet outputs a set of quotes in CSV format.
func (o Output) WriteSet(set [][]app.Quote) (map[int][]string, error) {
	w, closeWriter, err := o.openWriter()
	if err != nil {
		return nil, err
	}
	defer closeWriter()

	size := len(o.Symbols)
	rows, nextIdx := o.prepare(size)
	missing := make(map[int][]string)
	for i, quotes := range set {
		var (
			m   []string
			err error
		)
		m, err = o.bufferRows(quotes, rows, nextIdx)
		if err != nil {
			return missing, err
		}
		if len(m) > 0 {
			missing[i] = m
		}
		if err := o.writeBuffer(w, rows); err != nil {
			return missing, err
		}
		rows = make([][]string, size)
		nextIdx = 0
	}

	return missing, nil
}

// Write outputs a list of quotes in CSV format.
func (o Output) Write(quotes []app.Quote) (missing []string, err error) {
	rows, startIdx := o.prepare(len(o.Symbols))

	missing, err = o.bufferRows(quotes, rows, startIdx)
	if err != nil {
		return missing, err
	}

	w, closeWriter, err := o.openWriter()
	if err != nil {
		return missing, err
	}
	defer closeWriter()

	if err := o.writeBuffer(w, rows); err != nil {
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

func (o Output) bufferRows(quotes []app.Quote, rows [][]string, idx int) (missing []string, err error) {
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

func (o Output) openWriter() (Writer, func(), error) {
	f, err := CreateFile(o.Filename)
	if err != nil {
		return nil, nil, err
	}
	return NewWriter(f), func() { _ = f.Close }, nil
}

func (o Output) writeBuffer(w Writer, rows [][]string) error {
	for _, row := range rows {
		if len(row) == 0 {
			continue
		}
		if err := w.Write(row); err != nil {
			return fmt.Errorf("error writing out CSV row '%s': %v", o.Filename, err)
		}
	}
	w.Flush()

	return nil
}
