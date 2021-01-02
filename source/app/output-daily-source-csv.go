package app

import (
	"context"
)

// OutputDailySourceCSV retrieves the daily prices for the source and outputs to CSV file.
func (a App) OutputDailySourceCSV(ctx context.Context) error {
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

	// TODO: use provider to decode cached JSON into a collection of 'Coin' entities

	// TODO: store Coin entities in csv repo implementation (for importing into GnuCash)

	return nil
}
