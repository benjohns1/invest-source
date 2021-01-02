package main

import (
	"context"
	"github.com/benjohns1/invest-source/app"
	"github.com/benjohns1/invest-source/cache/file"
	"github.com/benjohns1/invest-source/provider/coinmarketcap"
	"github.com/spf13/viper"
	"log"
	"os"
)

type config struct {
	CoinMarketCapApiKey string
	CacheDirectory string
}

func parseCfg() config {
	viper.SetDefault("CacheDirectory", "./data")

	readCfgFile("ConfigFile", "config.yaml")
	readCfgFile("SecretConfigFile", ".secrets.yaml")

	cfg := config{}
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatal(err)
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

	log.Println("injecting dependencies")
	c, err := file.NewDailyCache(cfg.CacheDirectory)
	if err != nil {
		log.Fatal(err)
	}
	p, err := coinmarketcap.NewCoinMarketCapProvider(cfg.CoinMarketCapApiKey)
	if err != nil {
		log.Fatal(err)
	}
	a := app.App{
		Provider: p,
		Cache: c,
		Log: log.New(os.Stdout, "app: ", log.LstdFlags),
	}

	log.Println("outputting daily source CSV")
	if err := a.OutputDailySourceCSV(context.Background()); err != nil {
		log.Fatal(err)
	}

	log.Println("complete")
}
