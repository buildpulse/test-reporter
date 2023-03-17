package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/buildpulse/test-reporter/internal/cmd/submit"
	"github.com/buildpulse/test-reporter/internal/logger"
	"github.com/buildpulse/test-reporter/internal/metadata"
)

// set at buildtime via ldflags
var (
	Version = "development"
	Commit  = "unknown"
)

var usage = strings.ReplaceAll(`
CLI to submit test results to BuildPulse

USAGE
	$ %s submit TEST_RESULTS_PATH --account-id=ACCOUNT_ID --repository-id=REPOSITORY_ID

FLAGS
  --account-id      (required) BuildPulse account ID for the account that owns the repository
  --repository-id   (required) BuildPulse repository ID for the repository that produced the test results
  --repository-dir  Path to local git clone of the repository (default: ".")
  --tree            SHA-1 hash of the git tree that produced the test results (for use only if a local git clone does not exist)
	--endpoint-url    Hostname to point AWS client for uploading build archive

ENVIRONMENT VARIABLES
	Set the following environment variables:

	BUILDPULSE_ACCESS_KEY_ID      BuildPulse access key ID for the account that owns the repository

	BUILDPULSE_SECRET_ACCESS_KEY  BuildPulse secret access key for the account that owns the repository

	BUILDPULSE_BUCKET 						AWS S3 bucket for storing build archive (default: 'buildpulse-uploads')

EXAMPLE
	$ %s submit test/reports/*.xml --account-id 42 --repository-id 8675309 +-endpoint-url
`, "\t", "  ")

func main() {
	help := flag.Bool("help", false, "")
	version := flag.Bool("version", false, "")
	flag.Usage = func() {
		binaryName := os.Args[0]
		fmt.Fprintf(flag.CommandLine.Output(), usage, binaryName, binaryName)
	}
	flag.Parse()

	// if no cmd entered then exit
	if len(os.Args) == 1 {
		flag.Usage()
		os.Exit(1)
	}

	switch {
	case *help || os.Args[1] == "help":
		flag.Usage()
	case *version || os.Args[1] == "version":
		fmt.Print(getVersion().String())
	case os.Args[1] == "submit" && len(os.Args) > 2:
		log := logger.New(os.Stdout)
		c := submit.NewSubmit(getVersion(), log)
		envs := toMap(os.Environ())

		// validate args + env vars
		if err := c.Init(os.Args[2:], envs, submit.NewCommitResolverFactory(log)); err != nil {
			fmt.Fprintf(os.Stderr, "\n%s\n\nSee more help with --help\n", err)
			os.Exit(1)
		}
		_, err := c.Run() // run
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		flag.Usage()
		os.Exit(1)
	}

	os.Exit(0)
}

func toMap(pairs []string) map[string]string {
	m := map[string]string{}
	for _, s := range pairs {
		pair := strings.SplitN(s, "=", 2)
		m[pair[0]] = pair[1]
	}
	return m
}

func getVersion() *metadata.Version {
	return &metadata.Version{
		Commit:    Commit,
		GoOS:      runtime.GOOS,
		GoVersion: runtime.Version(),
		Number:    Version,
	}
}
