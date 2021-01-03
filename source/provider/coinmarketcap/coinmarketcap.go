package coinmarketcap

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

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

// ParseLatestQuotes
func (p Provider) ParseQuotes(_ []byte) ([]app.Quote, error) {
	return nil, fmt.Errorf("not implemented")
}
