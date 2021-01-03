package csv

import (
	"encoding/csv"
	"io"
	"os"
	"time"

	"github.com/benjohns1/invest-source/utils/filesystem"
)

var (
	// CreateFile for creating a local file.
	CreateFile = func(name string) (io.WriteCloser, error) { return os.Create(name) }

	// NewWriter creates a new CSV writer.
	NewWriter = func(w io.WriteCloser) Writer { return csv.NewWriter(w) }

	// Now implementation.
	Now = time.Now

	// Mkdir makes a directory if it doesn't exist.
	Mkdir = filesystem.Mkdir
)

// Writer CSV writing interface.
type Writer interface {
	Write([]string) error
	Flush()
}
