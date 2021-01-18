// +build mage

package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/benjohns1/invest-source/utils/filesystem"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	binary           = "bin/coinmarketcap-to-csv"
	src              = "cmd/coinmarketcap-to-csv/main.go"
	coverDir         = "coverage"
	packagePrefixLen = len("github.com/benjohns1/invest-source/")

	pullLambdaSrc    = "cmd/coinmarketcap-pull-aws-lambda/main.go"
	pullLambdaBinary = "build/artifacts/coinmarketcap-pull-aws-lambda"
	pullLambdaZip    = "build/artifacts/coinmarketcap-pull-aws-lambda.zip"
)

//goland:noinspection GoUnusedGlobalVariable
var Default = Start

// Install installs required tooling.
//goland:noinspection GoUnusedExportedFunction
func Install() error {
	if err := cmd("go", "get", "-u", "github.com/aws/aws-lambda-go/cmd/build-lambda-zip"); err != nil {
		return err
	}

	return nil
}

// Build the app for the current OS runtime.
func Build() error {
	binary := getBinaryForOS()
	if err := cmd("go", "build", "-o", binary, src); err != nil {
		return err
	}
	if err := envVars(map[string]string{"GOOS": "linux", "GOARCH": "amd64"}).cmd("go", "build", "-o", pullLambdaBinary, pullLambdaSrc); err != nil {
		return err
	}
	if err := zip(pullLambdaZip, pullLambdaBinary); err != nil {
		return err
	}

	return nil
}

// Start the app after building.
func Start() error {
	mg.Deps(Build)
	binary := getBinaryForOS()
	if err := cmd(binary, "--since=2021-01-01"); err != nil {
		return err
	}
	fmt.Printf("Ran %s\n", binary)
	return nil
}

// Test all packages within the app, generate coverage, and optionally open HTML coverage in the default browser.
//goland:noinspection GoUnusedExportedFunction
func Test(openInBrowser bool) error {
	if err := filesystem.Mkdir(coverDir); err != nil {
		return err
	}
	cover := fmt.Sprintf("%s/coverage.out", coverDir)
	coverHTML := fmt.Sprintf("%s/coverage.html", coverDir)

	if err := cmd("go", "test", "-coverprofile="+cover, "-covermode=count", "./..."); err != nil {
		return err
	}
	if err := cmd("go", "tool", "cover", "-html="+cover, "-o", coverHTML); err != nil {
		return err
	}

	absCoverHTML, err := filepath.Abs(coverHTML)
	if err != nil {
		return err
	}

	if openInBrowser {
		if err := openDefaultBrowser("file://" + absCoverHTML); err != nil {
			return err
		}
	}

	return nil
}

func cmd(cmd string, args ...string) error {
	return environment{}.cmd(cmd, args...)
}

type environment struct {
	vars map[string]string
}

func envVars(vars map[string]string) environment {
	return environment{
		vars,
	}
}

func (e environment) cmd(cmd string, args ...string) error {
	_, err := sh.Exec(e.vars, os.Stdout, os.Stderr, cmd, args...)
	return err
}

// awaitLocalstack blocks and polls until localstack is available or a timeout occurs
func awaitLocalstack() error {
	var ok bool
	const sleepTime = 3 * time.Second
	const retries = 20
	fmt.Println("checking if localstack is ready...")
	for i := 1; i <= retries; i++ {
		resp, err := http.Get("http://localhost:4566/health")
		if err != nil {
			fmt.Printf("localstack not ready, retrying after %v, attempt %d/%d...\n", sleepTime, i, retries)
			time.Sleep(sleepTime)
			continue
		}
		fmt.Printf("localstack response code: %d\n", resp.StatusCode)
		ok = true
		break
	}
	if !ok {
		return fmt.Errorf("timed out connecting to localstack")
	}
	return nil
}

// AWSLocalClean spins down localstack container and removes state files.
func AWSLocalClean() error {
	if err := cmd("docker-compose", "-f", "deploy/aws-local/docker-compose.yml", "down"); err != nil {
		fmt.Printf("error spinning down docker containers: %v\n", err)
	}

	if err := rm("deploy/aws-local/tf/terraform.tfstate"); err != nil {
		fmt.Printf("error removing statefile: %v\n", err)
	}

	return nil
}

// AWSLocal spins up a local AWS environment via localstack.
//goland:noinspection GoUnusedExportedFunction
func AWSLocal() error {
	mg.Deps(AWSLocalClean, Build)
	if err := cmd("docker-compose", "-f", "deploy/aws-local/docker-compose.yml", "up", "-d"); err != nil {
		return err
	}
	if err := awaitLocalstack(); err != nil {
		return err
	}

	const tfDirArg = "-chdir=deploy/aws-local/tf"
	const tfCmd = "terraform"
	if err := cmd(tfCmd, tfDirArg, "init"); err != nil {
		return err
	}
	if err := cmd(tfCmd, tfDirArg, "plan", "-out=.tfplan"); err != nil {
		return err
	}
	if err := cmd(tfCmd, tfDirArg, "apply", ".tfplan"); err != nil {
		return err
	}

	fmt.Println("AWS local started.")

	return nil
}

func zip(zip, src string) error {
	if runtime.GOOS == "windows" {
		return cmd("build-lambda-zip.exe", "-o", zip, src)
	}

	return cmd("zip", zip, src)
}

func rm(filepath string) error {
	var args []string
	switch runtime.GOOS {
	case "windows":
		winFile := strings.ReplaceAll(filepath, "/", "\\")
		fmt.Printf("removng windows file: %s\n", winFile)
		args = []string{"cmd", "/c", "del", winFile}
	default:
		fmt.Printf("removng file: %s\n", filepath)
		args = []string{"rm", filepath}
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Wait()
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
//goland:noinspection GoUnusedExportedFunction
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
