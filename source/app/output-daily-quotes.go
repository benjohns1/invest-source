package app

import "context"

// OutputDailyQuotes outputs the daily quotes since the last output, using cached source data.
func (a App) OutputDailyQuotes(_ context.Context) error {
	lastRun := a.Output.LastRun()

	set, err := a.Cache.ReadSince(lastRun)
	if err != nil {
		return err
	}

	a.Log.Printf("retrieved %d entries of cached data", len(set))

	quotes := make([][]Quote, len(set))
	for i, data := range set {
		var err error
		if quotes[i], err = a.Provider.ParseQuotes(data); err != nil {
			return err
		}
	}

	a.Log.Printf("writing output")

	missing, err := a.Output.WriteSet(quotes)
	if len(missing) > 0 {
		a.Log.Printf("missing symbols from output: %v", missing)
	}

	return err
}
