package keyval

import (
	"fmt"
	"strings"
	"time"
)

// Cache key-value store implementation.
type Cache struct {
	Provider Provider
	Bucket   string
	Key      func(int) string
}

// OldestCacheDate ...
var OldestCacheDate = time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC)

// Uploader for uploading files to a key-value store.
type Provider interface {
	Upload(bucket, key string, value []byte) error
	Download(bucket, key string) ([]byte, error)
}

// NewDailyCache instantiates a daily cache.
func NewDailyCache(provider Provider, bucket, pathPrefix string) (Cache, error) {
	c := Cache{
		Provider: provider,
		Bucket:   bucket,
		Key:      KeyGen(pathPrefix),
	}
	if err := c.Validate(); err != nil {
		return Cache{}, err
	}

	return c, nil
}

// Validate returns an error if the cache was not correctly instantiated.
func (c Cache) Validate() error {
	if c.Bucket == "" {
		return fmt.Errorf("cache keyval Bucket must be set")
	}
	if c.Key == nil {
		return fmt.Errorf("cache keyval Key must be set")
	}
	if c.Provider == nil {
		return fmt.Errorf("cache keyval Provider must be set")
	}

	return nil
}

// ReadCurrent retrieves the current day's cache data, or nil if it doesn't exist.
func (c Cache) ReadCurrent() ([]byte, error) {
	return c.Provider.Download(c.Bucket, c.Key(0))
}

// ReadSince retrieves all caches since the given time.
func (c Cache) ReadSince(since time.Time) ([][]byte, error) {
	if since.Before(OldestCacheDate) {
		since = OldestCacheDate
	}
	since = since.AddDate(0, 0, -1) // offset -1 day to be inclusive
	curr := Now().UTC()
	var set [][]byte
	for i := 0; ; i-- {
		curr = curr.AddDate(0, 0, -1)
		if curr.Before(since) {
			break
		}
		data, err := c.Provider.Download(c.Bucket, c.Key(i))
		if err != nil {
			return nil, err
		}
		if data == nil {
			continue
		}
		set = append(set, data)
	}
	return set, nil
}

// Write writes the data to a daily cache.
func (c Cache) WriteCurrent(data []byte) error {
	if err := c.Provider.Upload(c.Bucket, c.Key(0), data); err != nil {
		return err
	}

	return nil
}

// KeyGen returns a function to generate cache key names.
func KeyGen(path string) func(int) string {
	dirPath := strings.ReplaceAll(path, "\\", "/")
	if dirPath != "" && !strings.HasSuffix(dirPath, "/") {
		dirPath = dirPath + "/"
	}

	return func(dayOffset int) string {
		date := Now().UTC()
		if dayOffset != 0 {
			date = date.AddDate(0, 0, dayOffset)
		}
		return fmt.Sprintf("%s%s.json", dirPath, date.Format("2006-01-02"))
	}
}
