package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/benjohns1/invest-source/app"
	"github.com/benjohns1/invest-source/cache/file"
	"github.com/benjohns1/invest-source/output/csv"
	"github.com/benjohns1/invest-source/provider/coinmarketcap"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type config struct {
	CoinMarketCapApiKey string
	CacheDirectory      string
	OutputDirectory     string
	OutputSymbols       []string
	Since               string
}

func parseCfg() config {
	pflag.String("since", "2021-01-01", "output quote data since this date")
	pflag.Parse()
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		log.Fatal(err)
	}

	viper.SetDefault("CacheDirectory", "./data/cache")
	viper.SetDefault("OutputDirectory", "./data/out")
	viper.SetDefault("Since", "2021-01-01")

	readCfgFile("ConfigFile", "config.yaml")
	readCfgFile("SecretConfigFile", ".secrets.yaml")

	cfg := config{}
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatal(err)
	}
	for i, symbol := range cfg.OutputSymbols {
		cfg.OutputSymbols[i] = strings.TrimSpace(symbol)
	}

	log.Printf("parsed configs: %#v", cfg)
	return cfg
}

func readCfgFile(key string, defaultFile string) {
	viper.SetDefault(key, defaultFile)
	cfgFile := viper.GetString(key)
	log.Printf("reading %s from '%s'\n", key, cfgFile)
	viper.SetConfigFile(cfgFile)
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("unable to read %s file '%s', continuing with defaults: %v", key, cfgFile, err)
	}
}

func main() {
	log.Println("parsing config")
	cfg := parseCfg()

	ctx := context.Background()

	log.Println("injecting dependencies")
	c, err := file.NewDailyCache(cfg.CacheDirectory)
	if err != nil {
		log.Fatal(err)
	}
	p, err := coinmarketcap.NewCoinMarketCapProvider(cfg.CoinMarketCapApiKey)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("symbols to output: %v\n", cfg.OutputSymbols)
	o, err := csv.NewGnuCashCSV(cfg.OutputDirectory)
	if err != nil {
		log.Fatal(err)
	}
	a := app.App{
		Provider: p,
		Cache:    c,
		Output:   o,
		Log:      log.New(os.Stdout, "app: ", log.LstdFlags),
	}

	log.Println("caching daily source data")
	if err := a.CacheDailySourceData(ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("outputting daily quotes")
	if err := a.OutputDailyQuotes(ctx, cfg.Since, cfg.OutputSymbols); err != nil {
		log.Fatal(err)
	}

	log.Println("complete")
}
