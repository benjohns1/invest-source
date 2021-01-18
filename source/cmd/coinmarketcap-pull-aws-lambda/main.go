package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/benjohns1/invest-source/app"
	"github.com/benjohns1/invest-source/cache/keyval"
	keyvalProvider "github.com/benjohns1/invest-source/cache/keyval/provider"
	"github.com/benjohns1/invest-source/provider/coinmarketcap"
)

func main() {
	a, err := createApp()
	if err != nil {
		lambda.Start(func(ctx context.Context) error {
			return fmt.Errorf("error constructing lambda handler, request %s: %v", getAWSMeta(ctx).AwsRequestID, err)
		})
		return
	}
	lambda.Start(a.handleRequest)
}

func getAWSMeta(ctx context.Context) lambdacontext.LambdaContext {
	meta, _ := lambdacontext.FromContext(ctx)
	if meta == nil {
		log.Println("lambda context not found, using default")
		meta = &lambdacontext.LambdaContext{}
	}
	return *meta
}

type application struct {
	cfg config
}

// Provider ...
func (a application) Provider() app.Provider { return a.cfg.Provider }

// Cache ...
func (a application) Cache() app.Cache { return a.cfg.Cache }

// Log ...
func (a application) Log() app.Log { return a.cfg.Log }

func (a application) handleRequest(ctx context.Context) error {
	meta := getAWSMeta(ctx)
	log.Printf("request %s started", meta.AwsRequestID)
	defer log.Printf("request %s complete", meta.AwsRequestID)
	return app.CacheDailySourceData(ctx, a)
}

func createApp() (application, error) {
	log.Println("parsing config")
	cfg := parseCfg()

	log.Println("injecting dependencies")
	p, err := coinmarketcap.NewCoinMarketCapProvider(cfg.CoinMarketCapApiKey)
	if err != nil {
		return application{}, err
	}

	sess, err := session.NewSession(&aws.Config{
		Endpoint:         aws.String(cfg.AWSEndpoint),
		S3ForcePathStyle: aws.Bool(true),
		Region:           aws.String(cfg.AWSRegion),
	})
	if err != nil {
		return application{}, fmt.Errorf("error creating AWS session: %v", err)
	}

	s3, err := keyvalProvider.NewS3(sess)
	c, err := keyval.NewDailyCache(s3, cfg.CacheS3Bucket, "")
	if err != nil {
		return application{}, err
	}

	cfg.Provider = p
	cfg.Cache = c
	cfg.Log = log.New(os.Stdout, "app: ", log.LstdFlags)

	return application{
		cfg: cfg,
	}, nil
}

type config struct {
	CoinMarketCapApiKey string
	AWSEndpoint         string
	AWSRegion           string
	CacheS3Bucket       string
	Provider            app.Provider
	Cache               app.Cache
	Log                 app.Log
}

func parseCfg() config {
	cfg := config{
		CoinMarketCapApiKey: os.Getenv("CoinMarketCapApiKey"),
		AWSEndpoint:         os.Getenv("AWSEndpoint"),
		AWSRegion:           os.Getenv("AWSRegion"),
		CacheS3Bucket:       os.Getenv("CacheS3Bucket"),
	}

	log.Printf("parsed configs: %#v", cfg)
	return cfg
}
