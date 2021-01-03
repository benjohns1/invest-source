package coinmarketcap_test

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/benjohns1/invest-source/app"
	"github.com/benjohns1/invest-source/provider/coinmarketcap"
)

func TestProvider_ParseQuotes(t *testing.T) {
	type args struct {
		data    []byte
		symbols []string
	}
	tests := []struct {
		name     string
		provider *coinmarketcap.Provider
		args     args
		want     []app.Quote
		wantErr  bool
	}{
		{
			name: "should fail with invalid json",
			args: args{
				data: []byte("invalid-json"),
			},
			wantErr: true,
		},
		{
			name: "should fail with nil data",
			args: args{
				data: nil,
			},
			wantErr: true,
		},
		{
			name: "should return an empty array, given no data",
			args: args{
				data: []byte("{}"),
			},
			want: []app.Quote{},
		},
		{
			name: "should return a parsed array with a single quote, given valid data",
			args: args{
				data: []byte(`{
	"data": [
		{
			"symbol": "BTC",
			"quote": {
				"USD": {
					"price": 123456789.123456789,
					"last_updated": "2006-01-02T15:04:05.000Z"
				}
			}
		}
	]
}`),
			},
			want: []app.Quote{
				{
					Time:   time.Date(2006, time.January, 2, 15, 4, 5, 0, time.UTC),
					Symbol: "BTC",
					USD: func() decimal.Decimal {
						n, _ := decimal.NewFromString("123456789.123456789")
						return n
					}(),
				},
			},
		},
		{
			name: "should filter out data, if given a list of symbols",
			args: args{
				data: []byte(`{
	"data": [
		{
			"symbol": "BTC",
			"quote": {
				"USD": {
					"price": 123456789.123456789,
					"last_updated": "2006-01-02T15:04:05.000Z"
				}
			}
		}
	]
}`),
				symbols: []string{"NOT-BTC"},
			},
			want: []app.Quote{},
		},
		{
			name: "should fail if a symbol's price cannot be parsed",
			args: args{
				data: []byte(`{
	"data": [
		{
			"symbol": "BTC",
			"quote": {
				"USD": {
					"price": "invalid-json-number",
					"last_updated": "2006-01-02T15:04:05.000Z"
				}
			}
		}
	]
}`),
			},
			wantErr: true,
		},
		{
			name: "should fail if a symbol's last updated date cannot be parsed",
			args: args{
				data: []byte(`{
	"data": [
		{
			"symbol": "BTC",
			"quote": {
				"USD": {
					"price": 123456789.123456789,
					"last_updated": "invalid-date-format"
				}
			}
		}
	]
}`),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.provider == nil {
				p, err := coinmarketcap.NewCoinMarketCapProvider("dummy-api-key")
				if err != nil {
					t.Fatal(err)
				}
				tt.provider = &p
			}
			got, err := tt.provider.ParseQuotes(tt.args.data, tt.args.symbols...)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
