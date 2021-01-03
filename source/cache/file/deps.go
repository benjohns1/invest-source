package file

import (
	"github.com/benjohns1/invest-source/utils/filesystem"
	"io"
	"os"
	"time"
)

var (
	// Now function for retrieving the current timestamp. Override this for unit tests.
	Now = time.Now

	// OpenForReading opens a file for reading.
	OpenForReading = func(filename string) (io.Reader, error) { return os.Open(filename) }

	// CreateFile for creating a local file.
	CreateFile = func(name string) (io.Writer, error) { return os.Create(name) }

	// Mkdir makes a directory if it doesn't exist.
	Mkdir = filesystem.Mkdir
)