package csv

import (
	"fmt"

	"github.com/benjohns1/invest-source/app"
)

// CSV output implementation.
type Output struct {
	Dir       string
	HeaderRow []string
	Filter    func(app.Quote) bool
	MapRow    func(app.Quote) ([]string, error)
}

var DateFormat = "2006-01-02"

// NewGnuCashCSV outputs a CSV formatted for a GnuCash price import.
func NewGnuCashCSV(dir string) (Output, error) {
	if err := Mkdir(dir); err != nil {
		return Output{}, err
	}
	return Output{
		Dir:       dir,
		HeaderRow: []string{"Namespace", "Symbol", "Date", "Price", "Currency"},
		MapRow: func(q app.Quote) ([]string, error) {
			return []string{"AMEX", q.Symbol, q.Time.Format(DateFormat), q.USD.String(), "USD"}, nil
		},
	}, nil
}

// WriteSet outputs a set of quotes in CSV format.
func (o Output) WriteSet(filename string, set [][]app.Quote, symbols ...string) (map[int][]string, error) {
	w, closeWriter, err := o.openWriter(filename)
	if err != nil {
		return nil, err
	}
	defer closeWriter()

	rows := o.prepare()
	missing := make(map[int][]string)
	for i, quotes := range set {
		var (
			m   []string
			err error
		)
		rows, m, err = o.bufferRows(quotes, rows, symbols)
		if err != nil {
			return missing, err
		}
		if len(m) > 0 {
			missing[i] = m
		}
		if err := o.writeBuffer(w, rows); err != nil {
			return missing, err
		}
		rows = make([][]string, 0)
	}

	return missing, nil
}

func (o Output) prepare() [][]string {
	var rows [][]string
	if o.HeaderRow != nil {
		rows = make([][]string, 1)
		rows[0] = o.HeaderRow
	} else {
		rows = make([][]string, 0)
	}
	return rows
}

func (o Output) bufferRows(quotes []app.Quote, rows [][]string, symbols []string) (outRows [][]string, missing []string, err error) {
	found := make(map[string]struct{})
	for qNum, q := range quotes {
		if o.Filter != nil && !o.Filter(q) {
			continue
		}
		found[q.Symbol] = struct{}{}
		row, err := o.MapRow(q)
		if err != nil {
			return rows, nil, fmt.Errorf("quote number %d error: %v", qNum, err)
		}
		rows = append(rows, row)
	}

	if symbols != nil {
		for _, symbol := range symbols {
			if _, ok := found[symbol]; !ok {
				missing = append(missing, symbol)
			}
		}
	}
	return rows, missing, nil
}

func (o Output) openWriter(filename string) (Writer, func(), error) {
	f, err := CreateFile(fmt.Sprintf("%s/%s", o.Dir, filename))
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
			return fmt.Errorf("error writing out CSV row: %v", err)
		}
	}
	w.Flush()

	return nil
}
