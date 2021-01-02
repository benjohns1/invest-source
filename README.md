# Invest Source
Query investment API sources for current pricing data, and save to a personal repository.

## Build Prerequisites
- [Go](https://golang.org/)
- [Mage](https://github.com/magefile/mage) (or instead of running `mage` in the scripts below, use `go run mage.go`)

## Build and run
```
cd source
mage
```

## Configure
Required configs can be set via environment variables or in a `source/.secrets.yaml` file:
- **CoinMarketCapApiKey** - your API key from [coinmarketcap.com](https://pro.coinmarketcap.com/)

## Test
```
cd source
mage -v test 0
```
Open with HTML test coverage displayed in your default browser: `mage -v test 1`