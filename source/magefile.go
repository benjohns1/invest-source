// +build mage

package main

import (
	"fmt"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	binary = "bin/coinmarketcap-to-csv"
	src = "cmd/coinmarketcap-to-csv/main.go"
	coverDir = "coverage"
	packagePrefixLen = len("github.com/benjohns1/invest-source/")
)

var Default = Start

// Build the app for the current OS runtime.
func Build() error {
	binary :=  getBinaryForOS()
	return sh.Run("go", "build", "-o", binary, src)
}

// Start the app after building.
func Start() error {
	mg.Deps(Build)
	binary := getBinaryForOS()
	ran, err := sh.Exec(nil, os.Stdout, os.Stderr, binary)
	if err != nil {
		return err
	}
	if ran {
		log.Printf("Ran %s\n", binary)
	}

	return nil
}

// Test all packages within the app, generate coverage, and optionally open HTML coverage in the default browser.
func Test(openInBrowser bool) error {
	if err := mkdir(coverDir); err != nil {
		return err
	}
	cover := fmt.Sprintf("%s/coverage.out", coverDir)
	coverHTML := fmt.Sprintf("%s/coverage.html", coverDir)

	if err := sh.Run("go", "test", "-coverprofile="+cover, "-covermode=count", "./..."); err != nil {
		return err
	}
	if err := sh.Run("go", "tool", "cover", "-html="+cover, "-o", coverHTML); err != nil {
		return err
	}

	absCoverHTML, err := filepath.Abs(coverHTML)
	if err != nil {
		return err
	}

	if openInBrowser {
		if err := openDefaultBrowser("file://"+absCoverHTML); err != nil {
			return err
		}
	}

	return nil
}

func openDefaultBrowser(url string) error {
	var args []string
	switch runtime.GOOS {
	case "windows":
		args = []string{"cmd", "/c", "start"}
	case "darwin":
		args = []string{"open"}
	default:
		args = []string{"xdg-open"}
	}

	return exec.Command(args[0], append(args[1:], url)...).Start()
}

// List all Go packages in the app.
func List() error {
	pkgs, err := getPackages()
	if err != nil {
		return err
	}
	for _, p := range pkgs {
		fmt.Println(p.path)
	}
	return nil
}

func getBinaryForOS() string {
	if runtime.GOOS != "windows" {
		return binary
	}
	return fmt.Sprintf("%s.exe", binary)
}

type pkg struct {
	name string
	path string
}

func getPackages() ([]pkg, error) {
	s, err := sh.Output("go", "list", "./...")
	if err != nil {
		return nil, err
	}

	pstrs := strings.Split(s, "\n")
	pkgs := make([]pkg, len(pstrs))
	for i, pstr := range pstrs {
		path := pstr[packagePrefixLen:]
		pkgs[i].name = strings.ReplaceAll(strings.ReplaceAll(path, "/", "."), "\\", ".")
		pkgs[i].path = "./" + path
	}

	return pkgs, nil
}

func mkdir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("error attempting to create dir '%s': %v", dir, err)
		}
	} else if err != nil {
		return fmt.Errorf("error attempting to read dir '%s': %v", dir, err)
	}
	return nil
}