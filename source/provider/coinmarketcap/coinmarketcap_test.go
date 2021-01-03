package coinmarketcap_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/benjohns1/invest-source/app"
	"github.com/benjohns1/invest-source/provider/coinmarketcap"
)

func TestProvider_ParseQuotes(t *testing.T) {
	type args struct {
		data []byte
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
			got, err := tt.provider.ParseQuotes(tt.args.data)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
