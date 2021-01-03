package app

import (
	"context"
)

// OutputDailySourceCSV retrieves the daily prices for the source and outputs to CSV file.
func (a App) OutputDailySourceCSV(_ context.Context) error {
	data, err := a.Cache.ReadCurrent()
	if err != nil {
		return err
	}

	if data == nil {
		a.Log.Println("no daily cache file found, retrieving from API")

		var err error
		data, err = a.Provider.QueryLatest()
		if err != nil {
			return err
		}

		if err := a.Cache.WriteCurrent(data); err != nil {
			return err
		}
	} else {
		a.Log.Println("using daily cache file")
	}

	a.Log.Printf("retrieved %d bytes", len(data))

	quotes, err := a.Provider.ParseQuotes(data)
	if err != nil {
		return err
	}

	a.Log.Printf("writing output")

	missing, err := a.Output.Write(quotes)
	if len(missing) > 0 {
		a.Log.Printf("missing symbols from output: %v", missing)
	}

	return err
}
