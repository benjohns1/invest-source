package filesystem

import (
	"fmt"
	"os"
)

// Mkdir makes a directory.
func Mkdir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("error attempting to create dir '%s': %v", dir, err)
		}
	} else if err != nil {
		return fmt.Errorf("error attempting to read dir '%s': %v", dir, err)
	}
	return nil
}
