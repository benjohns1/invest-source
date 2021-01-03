package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

// Cache file implementation.
type Cache struct {
	Filename func(int) string
}

var oldestCacheDate = time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)

// NewDailyCache instantiates a daily cache.
func NewDailyCache(dir string) (Cache, error) {
	c := Cache{
		Filename: FilenameGen(dir),
	}
	if err := c.Validate(); err != nil {
		return Cache{}, err
	}
	return c, nil
}

// Validate returns an error if the cache was not correctly instantiated.
func (c Cache) Validate() error {
	if c.Filename == nil {
		return fmt.Errorf("cache Filename must be set")
	}

	return nil
}

// ReadCurrent retrieves the current day's cache file data, or nil if it doesn't exist.
func (c Cache) ReadCurrent() ([]byte, error) {
	return c.read(0)
}

func (c Cache) read(dayOffset int) ([]byte, error) {
	f, err := OpenForReading(c.Filename(dayOffset))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	return ioutil.ReadAll(f)
}

// ReadSince retrieves all caches since the given time.
func (c Cache) ReadSince(since time.Time) ([][]byte, error) {
	if since.Before(oldestCacheDate) {
		since = oldestCacheDate
	}
	curr := Now().UTC()
	var set [][]byte
	for i := 0; ; i-- {
		curr = curr.AddDate(0, 0, -1)
		if curr.Before(since) {
			break
		}
		data, err := c.read(i)
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
	f, err := CreateFile(c.Filename(0))
	if err != nil {
		return err
	}

	if _, err := f.Write(data); err != nil {
		return err
	}

	return nil
}

// FilenameGen returns a function to generate cache file names.
func FilenameGen(dir string) func(int) string {
	dirPath := strings.ReplaceAll(dir, "\\", "/")
	if dirPath != "" && !strings.HasSuffix(dirPath, "/") {
		dirPath = dirPath + "/"
	}
	_ = Mkdir(dirPath)

	return func(dayOffset int) string {
		date := Now().UTC()
		if dayOffset != 0 {
			date = date.AddDate(0, 0, dayOffset)
		}
		return fmt.Sprintf("%s%s.json", dirPath, date.Format("2006-01-02"))
	}
}
