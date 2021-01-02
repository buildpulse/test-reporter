package cmd

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buildpulse/cli/metadata"
	"github.com/dnaeon/go-vcr/cassette"
	"github.com/dnaeon/go-vcr/recorder"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var exampleEnv = map[string]string{
	"BUILDPULSE_ACCESS_KEY_ID":     "some-access-key-id",
	"BUILDPULSE_SECRET_ACCESS_KEY": "some-secret-access-key",
}

const (
	accessKeyID     = "REDACTED"
	secretAccessKey = "REDACTED"
)

func TestSubmit_Init(t *testing.T) {
	dir, err := os.Getwd()
	require.NoError(t, err)

	s := NewSubmit(&metadata.Version{})
	err = s.Init([]string{dir, "--account-id", "42", "--repository-id", "8675309"}, exampleEnv)
	assert.NoError(t, err)
	assert.Equal(t, dir, s.path)
	assert.EqualValues(t, 42, s.accountID)
	assert.EqualValues(t, 8675309, s.repositoryID)
	assert.Equal(t, "some-access-key-id", s.credentials.AccessKeyID)
	assert.Equal(t, "some-secret-access-key", s.credentials.SecretAccessKey)
	assert.Equal(t, exampleEnv, s.envs)
}

func TestSubmit_Init_invalidArgs(t *testing.T) {
	dir, err := os.Getwd()
	require.NoError(t, err)

	tests := []struct {
		name   string
		args   string
		errMsg string
	}{
		{
			name:   "PathWithNoFlags",
			args:   dir,
			errMsg: "missing required flag: -account-id",
		},
		{
			name:   "FlagsWithNoPath",
			args:   "--account-id 1 --repository-id 2",
			errMsg: "missing TEST_RESULTS_DIR",
		},
		{
			name:   "MissingAccountID",
			args:   fmt.Sprintf("%s --repository-id 2", dir),
			errMsg: "missing required flag: -account-id",
		},
		{
			name:   "MalformedAccountID",
			args:   fmt.Sprintf("%s --account-id bogus --repository-id 2", dir),
			errMsg: `invalid value "bogus" for flag -account-id`,
		},
		{
			name:   "MissingRepositoryID",
			errMsg: "missing required flag: -repository-id",
			args:   fmt.Sprintf("%s --account-id 1", dir),
		},
		{
			name:   "MalformedRepositoryID",
			args:   fmt.Sprintf("%s --account-id 1 --repository-id bogus", dir),
			errMsg: `invalid value "bogus" for flag -repository-id`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSubmit(&metadata.Version{})
			err := s.Init(strings.Split(tt.args, " "), exampleEnv)
			if assert.Error(t, err) {
				assert.Regexp(t, tt.errMsg, err.Error())
			}
		})
	}
}

func TestSubmit_Init_invalidEnv(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		errMsg  string
	}{
		{
			name: "MissingAccessKeyID",
			envVars: map[string]string{
				"BUILDPULSE_SECRET_ACCESS_KEY": "some-secret-access-key",
			},
			errMsg: "missing required environment variable: BUILDPULSE_ACCESS_KEY_ID",
		},
		{
			name: "EmptyAccessKeyID",
			envVars: map[string]string{
				"BUILDPULSE_ACCESS_KEY_ID":     "",
				"BUILDPULSE_SECRET_ACCESS_KEY": "some-secret-access-key",
			},
			errMsg: "missing required environment variable: BUILDPULSE_ACCESS_KEY_ID",
		},
		{
			name: "MissingSecretAccessKey",
			envVars: map[string]string{
				"BUILDPULSE_ACCESS_KEY_ID": "some-access-id",
			},
			errMsg: "missing required environment variable: BUILDPULSE_SECRET_ACCESS_KEY",
		},
		{
			name: "EmptySecretAccessKey",
			envVars: map[string]string{
				"BUILDPULSE_ACCESS_KEY_ID":     "some-access-id",
				"BUILDPULSE_SECRET_ACCESS_KEY": "",
			},
			errMsg: "missing required environment variable: BUILDPULSE_SECRET_ACCESS_KEY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, err := os.Getwd()
			require.NoError(t, err)

			s := NewSubmit(&metadata.Version{})
			err = s.Init([]string{dir, "--account-id", "42", "--repository-id", "8675309"}, tt.envVars)
			if assert.Error(t, err) {
				assert.Equal(t, tt.errMsg, err.Error())
			}
		})
	}
}

func TestSubmit_Init_invalidPath(t *testing.T) {
	t.Run("NonexistentPath", func(t *testing.T) {
		s := NewSubmit(&metadata.Version{})
		err := s.Init([]string{"some-nonexistent-path", "--account-id", "42", "--repository-id", "8675309"}, exampleEnv)
		if assert.Error(t, err) {
			assert.Equal(t, "path is not a directory: some-nonexistent-path", err.Error())
		}
	})

	t.Run("NonDirectoryPath", func(t *testing.T) {
		tmpfile, err := ioutil.TempFile("", "buildpulse-cli-test-fixture")
		require.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		s := NewSubmit(&metadata.Version{})
		err = s.Init([]string{tmpfile.Name(), "--account-id", "42", "--repository-id", "8675309"}, exampleEnv)
		if assert.Error(t, err) {
			assert.Regexp(t, "path is not a directory: ", err.Error())
		}
	})
}

func TestSubmit_Run(t *testing.T) {
	dir, err := ioutil.TempDir("", "buildpulse-cli-test-fixture")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	r, err := recorder.New("testdata/s3-success")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, r.Stop())
	}()

	envs := map[string]string{
		"GITHUB_ACTIONS": "true",
		"GITHUB_SHA":     "aaaaaaaaaaaaaaaaaaaabbbbbbbbbbbbbbbbbbbb",
	}

	s := &Submit{
		client:       &http.Client{Transport: r},
		idgen:        func() uuid.UUID { return uuid.MustParse("00000000-0000-0000-0000-000000000000") },
		version:      &metadata.Version{Number: "v1.2.3"},
		envs:         envs,
		path:         dir,
		accountID:    42,
		repositoryID: 8675309,
		credentials: credentials{
			AccessKeyID:     accessKeyID,
			SecretAccessKey: secretAccessKey,
		},
	}

	key, err := s.Run()
	require.NoError(t, err)

	yaml, err := ioutil.ReadFile(filepath.Join(dir, "buildpulse.yml"))
	require.NoError(t, err)
	assert.Contains(t, string(yaml), ":ci_provider: github-actions")
	assert.Contains(t, string(yaml), ":commit: aaaaaaaaaaaaaaaaaaaabbbbbbbbbbbbbbbbbbbb")
	assert.Contains(t, string(yaml), ":reporter_version: v1.2.3")

	assert.Equal(t, "8675309/buildpulse-00000000-0000-0000-0000-000000000000.gz", key)
}

func Test_upload(t *testing.T) {
	tests := []struct {
		name            string
		fixture         string
		accountID       uint64
		accessKeyID     string
		secretAccessKey string
		err             string
	}{
		{
			name:            "success",
			fixture:         "testdata/s3-success",
			accountID:       42,
			accessKeyID:     accessKeyID,
			secretAccessKey: secretAccessKey,
			err:             "",
		},
		{
			name:            "bad access key ID",
			fixture:         "testdata/s3-bad-access-key-id",
			accountID:       42,
			accessKeyID:     "some-bogus-access-key-id",
			secretAccessKey: secretAccessKey,
			err:             "InvalidAccessKeyId",
		},
		{
			name:            "bad secret access key",
			fixture:         "testdata/s3-bad-secret-access-key",
			accountID:       42,
			accessKeyID:     accessKeyID,
			secretAccessKey: "some-bogus-secret-access-key",
			err:             "SignatureDoesNotMatch",
		},
		{
			name:            "bad bucket",
			fixture:         "testdata/s3-bad-bucket",
			accountID:       1,
			accessKeyID:     accessKeyID,
			secretAccessKey: secretAccessKey,
			err:             "NoSuchBucket",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := recorder.New(tt.fixture)
			require.NoError(t, err)
			defer func() {
				require.NoError(t, r.Stop())
			}()

			r.SetMatcher(interactionMatcher)

			s := &Submit{
				client:       &http.Client{Transport: r},
				idgen:        func() uuid.UUID { return uuid.MustParse("00000000-0000-0000-0000-000000000000") },
				accountID:    tt.accountID,
				repositoryID: 8675309,
				credentials: credentials{
					AccessKeyID:     tt.accessKeyID,
					SecretAccessKey: tt.secretAccessKey,
				},
			}
			key, err := s.upload("testdata/example-test-results.tar.gz")
			if tt.err == "" {
				assert.NoError(t, err)
				assert.Equal(t, "8675309/buildpulse-00000000-0000-0000-0000-000000000000.gz", key)
			} else {
				assert.Contains(t, err.Error(), tt.err)
			}
		})
	}
}

func Test_toTarGz(t *testing.T) {
	path, err := toTarGz("./testdata/example-test-results")
	require.NoError(t, err)

	// === Unzip
	zipfile, err := os.Open(path)
	require.NoError(t, err)
	defer zipfile.Close()

	tarfile, err := ioutil.TempFile("", "buildpulse-unzip-")
	require.NoError(t, err)
	defer os.Remove(tarfile.Name())

	err = unzip(zipfile, tarfile)
	require.NoError(t, err)

	// === Untar
	dir, err := ioutil.TempDir("", "buildpulse-untar-")
	require.NoError(t, err)

	tarfile, err = os.Open(tarfile.Name())
	require.NoError(t, err)

	err = untar(tarfile, dir)
	require.NoError(t, err)

	// === Verify original directory content matches resulting directory content
	assertEqualContent(t,
		"testdata/example-test-results/buildpulse.yml",
		filepath.Join(dir, "buildpulse.yml"),
	)
	assertEqualContent(t,
		"testdata/example-test-results/junit/browserstack/example-1.xml",
		filepath.Join(dir, "junit/browserstack/example-1.xml"),
	)
	assertEqualContent(t,
		"testdata/example-test-results/junit/browserstack/example-2.XML",
		filepath.Join(dir, "junit/browserstack/example-2.XML"),
	)
	assertEqualContent(t,
		"testdata/example-test-results/junit/browsertest/example-3.xml",
		filepath.Join(dir, "junit/browsertest/example-3.xml"),
	)

	// === Verify tarball excludes files other than buildpulse.yml and XML reports
	assert.FileExists(t, "testdata/example-test-results/junit/browsertest/example-3.txt")
	assert.NoFileExists(t, filepath.Join(dir, "junit/browsertest/example-3.txt"))
}

func unzip(src io.Reader, dest io.Writer) error {
	zr, err := gzip.NewReader(src)
	if err != nil {
		return err
	}
	defer zr.Close()

	_, err = io.Copy(dest, zr)
	if err != nil {
		return err
	}

	return nil
}

func untar(src io.Reader, dest string) error {
	tarReader := tar.NewReader(src)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		path := filepath.Join(dest, header.Name)
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}
	}

	return nil
}

// assertEqualContent asserts that two files have the same content.
func assertEqualContent(t *testing.T, expected string, actual string) {
	expectedBytes, err := ioutil.ReadFile(expected)
	require.NoError(t, err)

	actualBytes, err := ioutil.ReadFile(actual)
	require.NoError(t, err)

	assert.Equal(t, expectedBytes, actualBytes)
}

// interactionMatcher provides a custom vcr matcher that compares the request
// method, URL, and body.
func interactionMatcher(r *http.Request, i cassette.Request) bool {
	if r.Body == nil {
		return i.Body == ""
	}
	var b bytes.Buffer
	if _, err := b.ReadFrom(r.Body); err != nil {
		return false
	}
	r.Body = ioutil.NopCloser(&b)
	return cassette.DefaultMatcher(r, i) && (b.String() == i.Body)
}
