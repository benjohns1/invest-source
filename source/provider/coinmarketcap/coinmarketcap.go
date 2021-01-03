package coinmarketcap

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/shopspring/decimal"

	"github.com/benjohns1/invest-source/app"
)

// Provider CoinMarketCap crypto API provider.
type Provider struct {
	ApiKey  string
	Limit   int
	Convert string
}

// NewCoinMarketCapProvider creates a new provider for the Coin Market Cap API (https://coinmarketcap.com/).
func NewCoinMarketCapProvider(apiKey string) (Provider, error) {
	p := Provider{
		ApiKey:  apiKey,
		Limit:   5000,
		Convert: "USD",
	}
	if err := p.Validate(); err != nil {
		return Provider{}, err
	}
	return p, nil
}

// Validate returns an error if the provider was not correctly instantiated.
func (p Provider) Validate() error {
	if p.ApiKey == "" {
		return fmt.Errorf("provider ApiKey must be set")
	}

	if p.Limit <= 0 {
		return fmt.Errorf("provider Limit must be greater than 0, got %d", p.Limit)
	}

	return nil
}

// QueryLatest retrieves the latest currency listing data from the CoinMarketCap API.
func (p Provider) QueryLatest() ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://pro-api.coinmarketcap.com/v1/cryptocurrency/listings/latest", nil)
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Add("limit", fmt.Sprintf("%d", p.Limit))
	q.Add("convert", p.Convert)

	req.Header.Set("Accepts", "application/json")
	req.Header.Add("X-CMC_PRO_API_KEY", p.ApiKey)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %s raw response body: %s", resp.Status, respBody)
	}

	return respBody, nil
}

type entry struct {
	Data []security `json:"data"`
}

type security struct {
	Symbol string `json:"symbol"`
	Quote  struct {
		USD struct {
			Price       json.Number `json:"price"`
			LastUpdated string      `json:"last_updated"`
		} `json:"USD"`
	} `json:"quote"`
}

// ParseLatestQuotes
func (p Provider) ParseQuotes(data []byte, symbols ...string) ([]app.Quote, error) {
	if data == nil {
		return nil, fmt.Errorf("data cannot be empty")
	}
	v := entry{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("error unmarshalling data into JSON: %v", err)
	}
	filterSymbols := len(symbols) > 0
	symbolMap := make(map[string]struct{}, len(symbols))
	for _, symbol := range symbols {
		symbolMap[symbol] = struct{}{}
	}
	quotes := make([]app.Quote, 0)
	for _, datum := range v.Data {
		if filterSymbols {
			if _, ok := symbolMap[datum.Symbol]; !ok {
				continue
			}
		}
		price, err := decimal.NewFromString(datum.Quote.USD.Price.String())
		if err != nil {
			return nil, fmt.Errorf("error parsing price for %s: %v", datum.Symbol, err)
		}
		t, err := time.Parse(time.RFC3339Nano, datum.Quote.USD.LastUpdated)
		if err != nil {
			return nil, fmt.Errorf("error parsing updated time for %s: %v", datum.Symbol, err)
		}
		quotes = append(quotes, app.Quote{
			Time:   t,
			Symbol: datum.Symbol,
			USD:    price,
		})
	}
	return quotes, nil
}
