package app

import (
	"context"
)

// CacheDailySourceData retrieves the daily prices for the source if it hasn't already, and caches the data.
func (a App) CacheDailySourceData(_ context.Context) error {
	data, err := a.Cache.ReadCurrent()
	if err != nil {
		return err
	}

	if data != nil {
		a.Log.Println("daily cache file found")
		return nil
	}

	a.Log.Println("no daily cache file found, retrieving from API")

	data, err = a.Provider.QueryLatest()
	if err != nil {
		return err
	}

	if err := a.Cache.WriteCurrent(data); err != nil {
		return err
	}

	a.Log.Printf("cached %d bytes", len(data))

	return nil
}
