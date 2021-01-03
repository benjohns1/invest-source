package app

import (
	"time"

	"github.com/shopspring/decimal"
)

// App provides methods for all application use-cases.
type App struct {
	Cache    Cache
	Provider Provider
	Output   Output
	Log      Log
}

// Quote contains a price quote for a single symbol at a point in time.
type Quote struct {
	Time   time.Time
	Symbol string
	USD    decimal.Decimal
}

// Cache caches API data when multiple use-cases are run for the same dataset without having to re-query the source API.
type Cache interface {
	ReadCurrent() ([]byte, error)
	WriteCurrent(data []byte) error
}

// Provider implements a source provider for retrieving external data.
type Provider interface {
	QueryLatest() ([]byte, error)
	ParseQuotes(data []byte) ([]Quote, error)
}

// Log interface.
type Log interface {
	Println(v ...interface{})
	Printf(format string, v ...interface{})
}

// Output implements an output writer.
type Output interface {
	Write([]Quote) ([]string, error)
}
