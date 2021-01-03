package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/benjohns1/invest-source/utils/filesystem"
)

// Now function for retrieving the current timestamp. Override this for unit tests.
var Now = time.Now

// Cache file implementation.
type Cache struct {
	CurrentFilename func() string
}

// NewDailyCache instantiates a daily cache.
func NewDailyCache(dir string) (Cache, error) {
	c := Cache{
		CurrentFilename: CurrentFilenameGen(dir),
	}
	if err := c.Validate(); err != nil {
		return Cache{}, err
	}
	return c, nil
}

// Validate returns an error if the cache was not correctly instantiated.
func (c Cache) Validate() error {
	if c.CurrentFilename == nil {
		return fmt.Errorf("cache CurrentFilename must be set")
	}

	return nil
}

// ReadCurrent retrieves the current day's cache file data, or nil if it doesn't exist.
func (c Cache) ReadCurrent() ([]byte, error) {
	f, err := os.Open(c.CurrentFilename())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	return ioutil.ReadAll(f)
}

// Write writes the data to a daily cache.
func (c Cache) WriteCurrent(data []byte) error {
	f, err := os.Create(c.CurrentFilename())
	if err != nil {
		return err
	}

	if _, err := f.Write(data); err != nil {
		return err
	}

	return nil
}

// CurrentFilenameGen returns a function to generate the current cache file name.
func CurrentFilenameGen(dir string) func() string {
	dirPath := strings.ReplaceAll(dir, "\\", "/")
	if dirPath != "" && !strings.HasSuffix(dirPath, "/") {
		dirPath = dirPath + "/"
	}
	_ = filesystem.Mkdir(dirPath)

	return func() string {
		return fmt.Sprintf("%s%s.json", dirPath, Now().UTC().Format("2006-01-02"))
	}
}
