package app

import (
	"context"
	"fmt"
	"time"
)

// DateFormat to use when parsing the 'since' datetime string.
var DateFormat = "2006-01-02"

// Now retrieves the current time.
var Now = time.Now

// OutputDailyQuotes outputs the daily quotes since the last output, using cached source data.
func (a App) OutputDailyQuotes(_ context.Context, since string, symbols []string) error {
	var sinceDate time.Time
	if since != "" {
		var err error
		if sinceDate, err = time.Parse(DateFormat, since); err != nil {
			return fmt.Errorf("error parsing 'since' date, should be of the form '%s', got '%s': %v", DateFormat, since, err)
		}
		sinceDate = sinceDate.UTC()
	}

	set, err := a.Cache.ReadSince(sinceDate)
	if err != nil {
		return err
	}

	a.Log.Printf("retrieved %d entries of cached data since %s", len(set), sinceDate.Format(DateFormat))

	quotes := make([][]Quote, len(set))
	for i, data := range set {
		var err error
		if quotes[i], err = a.Provider.ParseQuotes(data, symbols...); err != nil {
			return err
		}
	}

	a.Log.Printf("writing output")

	filename := fmt.Sprintf("%s_to_%s.csv", sinceDate.Format(DateFormat), Now().UTC().Format(DateFormat))
	missing, err := a.Output.WriteSet(filename, quotes, symbols...)
	if len(missing) > 0 {
		a.Log.Printf("missing symbols from output: %v", missing)
	}

	return err
}
