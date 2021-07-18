package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const binName = "buildpulse"

func TestMain(m *testing.M) {
	fmt.Println("Building...")
	build := exec.Command("go", "build", "-o", binName)
	if err := build.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot build %s: %s", binName, err)
		os.Exit(1)
	}

	fmt.Println("Running tests...")
	result := m.Run()

	fmt.Println("Cleaning up...")
	os.Remove(binName)

	os.Exit(result)
}

func TestCLI(t *testing.T) {
	dir, err := os.Getwd()
	require.NoError(t, err)
	cmdPath := filepath.Join(dir, binName)

	tests := []struct {
		name   string
		args   string
		errMsg string
		out    string
	}{
		{
			name:   "no subcommand",
			args:   "",
			errMsg: "exit status 1",
			out:    "USAGE",
		},
		{
			name:   "help subcommand",
			args:   "help",
			errMsg: "",
			out:    "USAGE",
		},
		{
			name:   "version subcommand",
			args:   "version",
			errMsg: "",
			out:    "BuildPulse Test Reporter development",
		},
		{
			name:   "version flag",
			args:   "--version",
			errMsg: "",
			out:    "BuildPulse Test Reporter development",
		},
		{
			name:   "submit subcommand without args",
			args:   "submit",
			errMsg: "exit status 1",
			out:    "USAGE",
		},
		{
			name:   "submit subcommand with invalid args",
			args:   "submit some-non-existent-path",
			errMsg: "exit status 1",
			out:    `no XML reports found at TEST_RESULTS_PATH`,
		},
		{
			name:   "unsupported subcommand",
			args:   "bogus",
			errMsg: "exit status 1",
			out:    "USAGE",
		},
		{
			name:   "unsupported flag",
			args:   "--bogus",
			errMsg: "exit status 2",
			out:    "USAGE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(cmdPath, strings.Split(tt.args, " ")...)
			out, err := cmd.CombinedOutput()
			assert.Contains(t, string(out), tt.out)
			if tt.errMsg != "" {
				if assert.Error(t, err) {
					assert.Equal(t, tt.errMsg, err.Error())
				}
			}
		})
	}
}
