package app

import (
	"context"
)

// CacheDailySourceDataDeps application dependencies for CacheDailySourceData use-case.
type CacheDailySourceDataDeps interface {
	Cache() Cache
	Provider() Provider
	Log() Log
}

// CacheDailySourceData retrieves the daily prices for the source if it hasn't already, and caches the data.
func CacheDailySourceData(_ context.Context, a CacheDailySourceDataDeps) error {
	data, err := a.Cache().ReadCurrent()
	if err != nil {
		return err
	}

	if data != nil {
		a.Log().Println("daily cache file found")
		return nil
	}

	a.Log().Println("no daily cache file found, retrieving from API")

	data, err = a.Provider().QueryLatest()
	if err != nil {
		return err
	}

	if err := a.Cache().WriteCurrent(data); err != nil {
		return err
	}

	a.Log().Printf("cached %d bytes", len(data))

	return nil
}
