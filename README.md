# Invest Source
Query investment API sources for current pricing data, and save to a personal repository.

## Build Prerequisites
- [Go](https://golang.org/)
- Docker Compose
- Terraform 0.14+
- [Mage](https://github.com/magefile/mage) (or instead of running `mage` in the scripts below, use `go run mage.go`)


## Configure
Required configs can be set via environment variables or in a `source/.secrets.yaml` file:
- **CoinMarketCapApiKey** - your API key from [coinmarketcap.com](https://pro.coinmarketcap.com/)
```
cd source
```

## Build and run
```
mage
```

## Test
Open new browser window with HTML test coverage:
```
mage -v test 1
```
Without opening browser:
```
mage -v test 0
```

## Run AWS infrastructure locally
Cache lambda will run every minute for testing.
```
mage awsLocal
```
After it has run, retrieve the cache file from local S3 (must have the AWS CLI installed):
```
aws --endpoint-url=http://localhost:4566 s3api get-object --bucket invest-source.coinmarketcap-pull-cache --key 2021-01-18.json output.json
```


Spin down:
```
mage awslocalclear
```