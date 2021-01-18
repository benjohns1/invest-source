package app

import (
	"time"

	"github.com/shopspring/decimal"
)

// App provides dependencies for all application use-cases.
type App struct {
	Config
}

// Config ...
type Config struct {
	Cache    Cache
	Provider Provider
	Output   Output
	Log      Log
}

// Cache ...
func (a App) Cache() Cache { return a.Config.Cache }

// Provider ...
func (a App) Provider() Provider { return a.Config.Provider }

// Output ...
func (a App) Output() Output { return a.Config.Output }

// Log ...
func (a App) Log() Log { return a.Config.Log }

// Quote contains a price quote for a single symbol at a point in time.
type Quote struct {
	Time   time.Time
	Symbol string
	USD    decimal.Decimal
}

// Cache caches API data when multiple use-cases are run for the same dataset without having to re-query the source API.
type Cache interface {
	ReadSince(time.Time) ([][]byte, error)
	ReadCurrent() ([]byte, error)
	WriteCurrent(data []byte) error
}

// Provider implements a source provider for retrieving external data.
type Provider interface {
	QueryLatest() ([]byte, error)
	ParseQuotes(data []byte, symbols ...string) ([]Quote, error)
}

// Log interface.
type Log interface {
	Println(v ...interface{})
	Printf(format string, v ...interface{})
}

// Output implements an output writer.
type Output interface {
	WriteSet(filename string, set [][]Quote, symbols ...string) (map[int][]string, error)
}
