package app

import "context"

// OutputDailyQuotes outputs the daily quotes since the last output, using cached source data.
func (a App) OutputDailyQuotes(_ context.Context) error {
	data, err := a.Cache.ReadCurrent()
	if err != nil {
		return err
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
