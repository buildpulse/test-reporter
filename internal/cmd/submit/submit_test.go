package submit

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buildpulse/test-reporter/internal/logger"
	"github.com/buildpulse/test-reporter/internal/metadata"
	"github.com/dnaeon/go-vcr/cassette"
	"github.com/dnaeon/go-vcr/recorder"
	"github.com/google/uuid"
	"github.com/mholt/archiver/v3"
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
	t.Run("MinimumRequiredArgs", func(t *testing.T) {
		s := NewSubmit(&metadata.Version{}, logger.New())
		err := s.Init([]string{"testdata/example-reports-dir/example-*.xml", "--account-id", "42", "--repository-id", "8675309"}, exampleEnv, new(stubCommitResolverFactory))
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"testdata/example-reports-dir/example-1.xml"}, s.paths)
		assert.EqualValues(t, 42, s.accountID)
		assert.EqualValues(t, 8675309, s.repositoryID)
		assert.Equal(t, "buildpulse-uploads", s.bucket)
		assert.Equal(t, "some-access-key-id", s.credentials.AccessKeyID)
		assert.Equal(t, "some-secret-access-key", s.credentials.SecretAccessKey)
		assert.Equal(t, exampleEnv, s.envs)
		assert.Equal(t, ".", s.repositoryPath)
		assert.Equal(t, "Repository", s.commitResolver.Source())
		assert.Equal(t, s.coveragePaths, []string{})
	})

	t.Run("WithCoveragePathString", func(t *testing.T) {
		s := NewSubmit(&metadata.Version{}, logger.New())
		err := s.Init([]string{"testdata/example-reports-dir/example-*.xml", "--account-id", "42", "--repository-id", "8675309", "--coverage-files", "./dir1/**/*.xml ./dir2/**/*.xml"}, exampleEnv, new(stubCommitResolverFactory))
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"testdata/example-reports-dir/example-1.xml"}, s.paths)
		assert.EqualValues(t, 42, s.accountID)
		assert.EqualValues(t, 8675309, s.repositoryID)
		assert.Equal(t, "buildpulse-uploads", s.bucket)
		assert.Equal(t, "some-access-key-id", s.credentials.AccessKeyID)
		assert.Equal(t, "some-secret-access-key", s.credentials.SecretAccessKey)
		assert.Equal(t, exampleEnv, s.envs)
		assert.Equal(t, ".", s.repositoryPath)
		assert.Equal(t, "Repository", s.commitResolver.Source())
		assert.Equal(t, s.coveragePaths, []string{"./dir1/**/*.xml", "./dir2/**/*.xml"})
	})

	t.Run("WithDisableCoverageAutoDiscovery", func(t *testing.T) {
		s := NewSubmit(&metadata.Version{}, logger.New())
		err := s.Init([]string{"testdata/example-reports-dir/example-*.xml", "--account-id", "42", "--repository-id", "8675309", "--disable-coverage-auto"}, exampleEnv, new(stubCommitResolverFactory))
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"testdata/example-reports-dir/example-1.xml"}, s.paths)
		assert.EqualValues(t, 42, s.accountID)
		assert.EqualValues(t, 8675309, s.repositoryID)
		assert.Equal(t, "buildpulse-uploads", s.bucket)
		assert.Equal(t, "some-access-key-id", s.credentials.AccessKeyID)
		assert.Equal(t, "some-secret-access-key", s.credentials.SecretAccessKey)
		assert.Equal(t, exampleEnv, s.envs)
		assert.Equal(t, ".", s.repositoryPath)
		assert.Equal(t, "Repository", s.commitResolver.Source())
		assert.True(t, s.disableCoverageAutoDiscovery)
	})

	t.Run("WithTagsString", func(t *testing.T) {
		s := NewSubmit(&metadata.Version{}, logger.New())
		err := s.Init([]string{"testdata/example-reports-dir/example-*.xml", "--account-id", "42", "--repository-id", "8675309", "--tags", "tag1 tag2"}, exampleEnv, new(stubCommitResolverFactory))
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"testdata/example-reports-dir/example-1.xml"}, s.paths)
		assert.EqualValues(t, 42, s.accountID)
		assert.EqualValues(t, 8675309, s.repositoryID)
		assert.Equal(t, "buildpulse-uploads", s.bucket)
		assert.Equal(t, "some-access-key-id", s.credentials.AccessKeyID)
		assert.Equal(t, "some-secret-access-key", s.credentials.SecretAccessKey)
		assert.Equal(t, exampleEnv, s.envs)
		assert.Equal(t, ".", s.repositoryPath)
		assert.Equal(t, "Repository", s.commitResolver.Source())
		assert.Equal(t, s.tagsString, "tag1 tag2")
	})

	t.Run("WithTagsString", func(t *testing.T) {
		s := NewSubmit(&metadata.Version{}, logger.New())
		err := s.Init([]string{"testdata/example-reports-dir/example-*.xml", "--account-id", "42", "--repository-id", "8675309", "--tags", "tag1 tag2", "--quota-id", "quota1"}, exampleEnv, new(stubCommitResolverFactory))
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"testdata/example-reports-dir/example-1.xml"}, s.paths)
		assert.EqualValues(t, 42, s.accountID)
		assert.EqualValues(t, 8675309, s.repositoryID)
		assert.Equal(t, "buildpulse-uploads", s.bucket)
		assert.Equal(t, "some-access-key-id", s.credentials.AccessKeyID)
		assert.Equal(t, "some-secret-access-key", s.credentials.SecretAccessKey)
		assert.Equal(t, "quota1", s.quotaID)
		assert.Equal(t, exampleEnv, s.envs)
		assert.Equal(t, ".", s.repositoryPath)
		assert.Equal(t, "Repository", s.commitResolver.Source())
		assert.Equal(t, s.tagsString, "tag1 tag2")
	})

	t.Run("WithMultiplePathArgs", func(t *testing.T) {
		s := NewSubmit(&metadata.Version{}, logger.New())
		err := s.Init(
			[]string{"testdata/example-reports-dir/example-1.xml", "testdata/example-reports-dir/example-2.XML", "--account-id", "42", "--repository-id", "8675309"},
			exampleEnv,
			new(stubCommitResolverFactory),
		)
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"testdata/example-reports-dir/example-1.xml", "testdata/example-reports-dir/example-2.XML"}, s.paths)
		assert.EqualValues(t, 42, s.accountID)
		assert.EqualValues(t, 8675309, s.repositoryID)
	})

	t.Run("WithDirectoryWithReportsAsPathArg", func(t *testing.T) {
		s := NewSubmit(&metadata.Version{}, logger.New())
		err := s.Init(
			[]string{"testdata/example-reports-dir/dir-with-xml-files/browserstack", "--account-id", "42", "--repository-id", "8675309"},
			exampleEnv,
			new(stubCommitResolverFactory),
		)
		require.NoError(t, err)
		assert.ElementsMatch(t,
			[]string{
				"testdata/example-reports-dir/dir-with-xml-files/browserstack/example-1.xml",
				"testdata/example-reports-dir/dir-with-xml-files/browserstack/example-2.xml",
			},
			s.paths,
		)
		assert.EqualValues(t, 42, s.accountID)
		assert.EqualValues(t, 8675309, s.repositoryID)
	})

	// To maintain backwards compatibility with releases prior to v0.19.0, if
	// exactly one path is given, and it's a directory, and it contains no XML
	// reports, then continue without erroring. The resulting upload will contain
	// *zero* XML reports.
	//
	// TODO: Treat this scenario as an error for the next major version release.
	t.Run("WithDirectoryWithoutReportsAsPathArg", func(t *testing.T) {
		s := NewSubmit(&metadata.Version{}, logger.New())
		err := s.Init(
			[]string{"testdata/example-reports-dir/dir-without-xml-files", "--account-id", "42", "--repository-id", "8675309"},
			exampleEnv,
			new(stubCommitResolverFactory),
		)
		require.NoError(t, err)
		assert.Empty(t, s.paths)
		assert.EqualValues(t, 42, s.accountID)
		assert.EqualValues(t, 8675309, s.repositoryID)
	})

	t.Run("WithRepositoryDirArg", func(t *testing.T) {
		repoDir := t.TempDir()

		s := NewSubmit(&metadata.Version{}, logger.New())
		err := s.Init(
			[]string{"testdata/example-reports-dir/example-*.xml", "--account-id", "42", "--repository-id", "8675309", "--repository-dir", repoDir},
			exampleEnv,
			new(stubCommitResolverFactory),
		)
		require.NoError(t, err)
		assert.Equal(t, repoDir, s.repositoryPath)
		assert.Equal(t, "Repository", s.commitResolver.Source())
	})

	t.Run("WithTreeArg", func(t *testing.T) {
		s := NewSubmit(&metadata.Version{}, logger.New())
		err := s.Init(
			[]string{"testdata/example-reports-dir/example-*.xml", "--account-id", "42", "--repository-id", "8675309", "--tree", "0000000000000000000000000000000000000000"},
			exampleEnv,
			new(stubCommitResolverFactory),
		)
		require.NoError(t, err)
		assert.Equal(t, "Static", s.commitResolver.Source())
	})

	t.Run("WithBuildPulseBucketEnvVar", func(t *testing.T) {
		repoDir := t.TempDir()

		envs := map[string]string{
			"BUILDPULSE_ACCESS_KEY_ID":     "some-access-key-id",
			"BUILDPULSE_SECRET_ACCESS_KEY": "some-secret-access-key",
			"BUILDPULSE_BUCKET":            "buildpulse-uploads-test",
		}
		s := NewSubmit(&metadata.Version{}, logger.New())
		err := s.Init(
			[]string{"testdata/example-reports-dir/example-*.xml", "--account-id", "42", "--repository-id", "8675309", "--repository-dir", repoDir},
			envs,
			new(stubCommitResolverFactory),
		)
		require.NoError(t, err)
		assert.Equal(t, "buildpulse-uploads-test", s.bucket)
	})
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
			errMsg: "missing TEST_RESULTS_PATH",
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
		{
			name:   "TreeFlagGivenWithMissingValue",
			args:   fmt.Sprintf("%s --account-id 1 --repository-id 2 --tree", dir),
			errMsg: `flag needs an argument: -tree`,
		},
		{
			name:   "TreeLengthInvalid",
			args:   fmt.Sprintf("%s --account-id 1 --repository-id 2 --tree abc", dir),
			errMsg: `invalid value "abc" for flag -tree: should be a 40-character SHA-1 hash`,
		},
		{
			name:   "TreeCharactersInvalid",
			args:   fmt.Sprintf("%s --account-id 1 --repository-id 2 --tree xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", dir),
			errMsg: `invalid value "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" for flag -tree: should be a 40-character SHA-1 hash`,
		},
		{
			name:   "TreeAndRepoPathBothGiven",
			args:   fmt.Sprintf("%s --account-id 1 --repository-id 2 --repository-dir . --tree 0000000000000000000000000000000000000000", dir),
			errMsg: `invalid use of flag -repository-dir with flag -tree: use one or the other, but not both`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSubmit(&metadata.Version{}, logger.New())
			err := s.Init(strings.Split(tt.args, " "), exampleEnv, &stubCommitResolverFactory{})
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

			s := NewSubmit(&metadata.Version{}, logger.New())
			err = s.Init([]string{dir, "--account-id", "42", "--repository-id", "8675309"}, tt.envVars, &stubCommitResolverFactory{})
			if assert.Error(t, err) {
				assert.Equal(t, tt.errMsg, err.Error())
			}
		})
	}
}

func TestSubmit_Init_invalidPaths(t *testing.T) {
	t.Run("NonexistentPath", func(t *testing.T) {
		s := NewSubmit(&metadata.Version{}, logger.New())
		err := s.Init([]string{
			"some-nonexistent-path",
			"--account-id", "42",
			"--repository-id", "8675309",
		},
			exampleEnv,
			&stubCommitResolverFactory{},
		)
		if assert.Error(t, err) {
			assert.Equal(t, "no XML reports found at TEST_RESULTS_PATH: some-nonexistent-path", err.Error())
		}
	})

	t.Run("MultiplePathsWithNoReportFiles", func(t *testing.T) {
		s := NewSubmit(&metadata.Version{}, logger.New())
		err := s.Init([]string{
			"testdata/example-reports-dir/dir-without-xml-files",
			"testdata/example-reports-dir/example.txt",
			"--account-id", "42",
			"--repository-id", "8675309",
		},
			exampleEnv,
			&stubCommitResolverFactory{},
		)
		if assert.Error(t, err) {
			assert.Equal(t, "no XML reports found at TEST_RESULTS_PATH: testdata/example-reports-dir/dir-without-xml-files testdata/example-reports-dir/example.txt", err.Error())
		}
	})

	t.Run("GlobPathMatchingNoReportFiles", func(t *testing.T) {
		s := NewSubmit(&metadata.Version{}, logger.New())
		err := s.Init([]string{"testdata/example-reports-dir/bogus*", "--account-id", "42", "--repository-id", "8675309"}, exampleEnv, &stubCommitResolverFactory{})
		if assert.Error(t, err) {
			assert.Equal(t, "no XML reports found at TEST_RESULTS_PATH: testdata/example-reports-dir/bogus*", err.Error())
		}
	})

	t.Run("BadGlobPattern", func(t *testing.T) {
		s := NewSubmit(&metadata.Version{}, logger.New())
		err := s.Init([]string{
			"[",
			"--account-id", "42",
			"--repository-id", "8675309",
		},
			exampleEnv,
			&stubCommitResolverFactory{},
		)
		if assert.Error(t, err) {
			assert.Equal(t, `invalid value "[" for path: syntax error in pattern`, err.Error())
		}
	})
}

func TestSubmit_Init_invalidRepoPath(t *testing.T) {
	t.Run("NonRepoPath", func(t *testing.T) {
		log := logger.New()
		s := NewSubmit(&metadata.Version{}, logger.New())
		err := s.Init([]string{
			".",
			"--account-id", "42",
			"--repository-id", "8675309",
			"--repository-dir", os.TempDir(),
		},
			exampleEnv,
			NewCommitResolverFactory(log),
		)
		if assert.Error(t, err) {
			assert.Regexp(t, "invalid value for flag -repository-dir: no repository found at ", err.Error())
		}
	})

	t.Run("NonexistentRepoPath", func(t *testing.T) {
		s := NewSubmit(&metadata.Version{}, logger.New())
		err := s.Init([]string{
			".",
			"--account-id", "42",
			"--repository-id", "8675309",
			"--repository-dir", filepath.Join(os.TempDir(), "some-nonexistent-path"),
		},
			exampleEnv,
			&stubCommitResolverFactory{},
		)
		if assert.Error(t, err) {
			assert.Regexp(t, "invalid value for flag -repository-dir: .* is not a directory", err.Error())
		}
	})

	t.Run("NonDirectoryRepoPath", func(t *testing.T) {
		tmpfile, err := os.CreateTemp(os.TempDir(), "buildpulse-cli-test-fixture")
		require.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		s := NewSubmit(&metadata.Version{}, logger.New())
		err = s.Init([]string{
			".",
			"--account-id", "42",
			"--repository-id", "8675309",
			"--repository-dir", tmpfile.Name(),
		},
			exampleEnv,
			&stubCommitResolverFactory{},
		)
		if assert.Error(t, err) {
			assert.Regexp(t, "invalid value for flag -repository-dir: .* is not a directory", err.Error())
		}
	})
}

func TestSubmit_Run(t *testing.T) {
	r, err := recorder.New("testdata/s3-success")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, r.Stop())
	}()

	envs := map[string]string{
		"GITHUB_ACTIONS": "true",
		"GITHUB_SHA":     "aaaaaaaaaaaaaaaaaaaabbbbbbbbbbbbbbbbbbbb",
	}

	log := logger.New()
	s := &Submit{
		client:         &http.Client{Transport: r},
		idgen:          func() uuid.UUID { return uuid.MustParse("00000000-0000-0000-0000-000000000000") },
		logger:         log,
		version:        &metadata.Version{Number: "v1.2.3"},
		commitResolver: metadata.NewStaticCommitResolver(&metadata.Commit{TreeSHA: "ccccccccccccccccccccdddddddddddddddddddd"}, log),
		envs:           envs,
		paths:          []string{"testdata/example-reports-dir/example-1.xml"},
		bucket:         "buildpulse-uploads",
		accountID:      42,
		repositoryID:   8675309,
		credentials: credentials{
			AccessKeyID:     accessKeyID,
			SecretAccessKey: secretAccessKey,
		},
	}

	key, err := s.Run()
	require.NoError(t, err)
	assert.Equal(t, "42/8675309/buildpulse-00000000-0000-0000-0000-000000000000.gz", key)
}

func Test_bundle(t *testing.T) {
	t.Run("bundle with coverage files provided", func(t *testing.T) {
		envs := map[string]string{
			"GITHUB_ACTIONS": "true",
			"GITHUB_SHA":     "aaaaaaaaaaaaaaaaaaaabbbbbbbbbbbbbbbbbbbb",
		}

		log := logger.New()
		s := &Submit{
			logger:         log,
			version:        &metadata.Version{Number: "v1.2.3"},
			commitResolver: metadata.NewStaticCommitResolver(&metadata.Commit{TreeSHA: "ccccccccccccccccccccdddddddddddddddddddd"}, log),
			envs:           envs,
			paths:          []string{"testdata/example-reports-dir/example-1.xml"},
			coveragePaths:  []string{"testdata/example-reports-dir/coverage/report.xml", "testdata/example-reports-dir/coverage/report-2.xml"},
			bucket:         "buildpulse-uploads",
			accountID:      42,
			repositoryID:   8675309,
		}

		path, err := s.bundle()
		require.NoError(t, err)

		unzipDir := t.TempDir()
		err = archiver.Unarchive(path, unzipDir)
		require.NoError(t, err)

		// Verify buildpulse.yml is present and contains expected content
		yaml, err := os.ReadFile(filepath.Join(unzipDir, "buildpulse.yml"))
		require.NoError(t, err)
		assert.Contains(t, string(yaml), ":ci_provider: github-actions")
		assert.Contains(t, string(yaml), ":commit: aaaaaaaaaaaaaaaaaaaabbbbbbbbbbbbbbbbbbbb")
		assert.Contains(t, string(yaml), ":tree: ccccccccccccccccccccdddddddddddddddddddd")
		assert.Contains(t, string(yaml), ":reporter_version: v1.2.3")

		// Verify test report XML file is present and contains expected content
		assertEqualContent(t,
			"testdata/example-reports-dir/example-1.xml",
			filepath.Join(unzipDir, "test_results/testdata/example-reports-dir/example-1.xml"),
		)

		// Verify coverage files are present and contains expected content
		assertEqualContent(t,
			"testdata/example-reports-dir/coverage/report.xml",
			filepath.Join(unzipDir, "coverage/testdata/example-reports-dir/coverage/report.xml"),
		)

		assertEqualContent(t,
			"testdata/example-reports-dir/coverage/report-2.xml",
			filepath.Join(unzipDir, "coverage/testdata/example-reports-dir/coverage/report-2.xml"),
		)

		// Verify buildpulse.log is present and contains expected content
		logdata, err := os.ReadFile(filepath.Join(unzipDir, "buildpulse.log"))
		require.NoError(t, err)
		assert.Contains(t, string(logdata), "Gathering metadata to describe the build")
	})

	t.Run("bundle with no coverage files provided (inferred)", func(t *testing.T) {
		envs := map[string]string{
			"GITHUB_ACTIONS": "true",
			"GITHUB_SHA":     "aaaaaaaaaaaaaaaaaaaabbbbbbbbbbbbbbbbbbbb",
		}

		log := logger.New()
		s := &Submit{
			logger:         log,
			version:        &metadata.Version{Number: "v1.2.3"},
			commitResolver: metadata.NewStaticCommitResolver(&metadata.Commit{TreeSHA: "ccccccccccccccccccccdddddddddddddddddddd"}, log),
			envs:           envs,
			paths:          []string{"testdata/example-reports-dir/example-1.xml"},
			bucket:         "buildpulse-uploads",
			accountID:      42,
			repositoryID:   8675309,
		}

		path, err := s.bundle()
		require.NoError(t, err)

		unzipDir := t.TempDir()
		err = archiver.Unarchive(path, unzipDir)
		require.NoError(t, err)

		// Verify buildpulse.yml is present and contains expected content
		yaml, err := os.ReadFile(filepath.Join(unzipDir, "buildpulse.yml"))
		require.NoError(t, err)
		assert.Contains(t, string(yaml), ":ci_provider: github-actions")
		assert.Contains(t, string(yaml), ":commit: aaaaaaaaaaaaaaaaaaaabbbbbbbbbbbbbbbbbbbb")
		assert.Contains(t, string(yaml), ":tree: ccccccccccccccccccccdddddddddddddddddddd")
		assert.Contains(t, string(yaml), ":reporter_version: v1.2.3")

		// Verify test report XML file is present and contains expected content
		assertEqualContent(t,
			"testdata/example-reports-dir/example-1.xml",
			filepath.Join(unzipDir, "test_results/testdata/example-reports-dir/example-1.xml"),
		)

		// Verify coverage file is present and contains expected content
		assertEqualContent(t,
			"testdata/example-reports-dir/coverage/report.xml",
			filepath.Join(unzipDir, "coverage/testdata/example-reports-dir/coverage/report.xml"),
		)

		ignoredCoverageReportPath := filepath.Join(unzipDir, "coverage/testdata/example-reports-dir/coverage/report-2.xml")
		_, err = os.Stat(ignoredCoverageReportPath)
		assert.True(t, os.IsNotExist(err))

		ignoredSourceFilePath := filepath.Join(unzipDir, "coverage/testdata/example-reports-dir/vendor/simplecov/coverage_statistic.go")
		_, err = os.Stat(ignoredSourceFilePath)
		assert.True(t, os.IsNotExist(err))

		// Verify buildpulse.log is present and contains expected content
		logdata, err := os.ReadFile(filepath.Join(unzipDir, "buildpulse.log"))
		require.NoError(t, err)
		assert.Contains(t, string(logdata), "Gathering metadata to describe the build")
	})

	t.Run("bundle with disabled coverage file autodetection", func(t *testing.T) {
		envs := map[string]string{
			"GITHUB_ACTIONS": "true",
			"GITHUB_SHA":     "aaaaaaaaaaaaaaaaaaaabbbbbbbbbbbbbbbbbbbb",
		}

		log := logger.New()
		s := &Submit{
			logger:                       log,
			version:                      &metadata.Version{Number: "v1.2.3"},
			commitResolver:               metadata.NewStaticCommitResolver(&metadata.Commit{TreeSHA: "ccccccccccccccccccccdddddddddddddddddddd"}, log),
			envs:                         envs,
			paths:                        []string{"testdata/example-reports-dir/example-1.xml"},
			bucket:                       "buildpulse-uploads",
			disableCoverageAutoDiscovery: true,
			accountID:                    42,
			repositoryID:                 8675309,
		}

		path, err := s.bundle()
		require.NoError(t, err)

		unzipDir := t.TempDir()
		err = archiver.Unarchive(path, unzipDir)
		require.NoError(t, err)

		// Verify buildpulse.yml is present and contains expected content
		yaml, err := os.ReadFile(filepath.Join(unzipDir, "buildpulse.yml"))
		require.NoError(t, err)
		assert.Contains(t, string(yaml), ":ci_provider: github-actions")
		assert.Contains(t, string(yaml), ":commit: aaaaaaaaaaaaaaaaaaaabbbbbbbbbbbbbbbbbbbb")
		assert.Contains(t, string(yaml), ":tree: ccccccccccccccccccccdddddddddddddddddddd")
		assert.Contains(t, string(yaml), ":reporter_version: v1.2.3")

		// Verify test report XML file is present and contains expected content
		assertEqualContent(t,
			"testdata/example-reports-dir/example-1.xml",
			filepath.Join(unzipDir, "test_results/testdata/example-reports-dir/example-1.xml"),
		)

		ignoredCoverageReportPath := filepath.Join(unzipDir, "coverage/testdata/example-reports-dir/coverage/report.xml")
		_, err = os.Stat(ignoredCoverageReportPath)
		assert.True(t, os.IsNotExist(err))

		// Verify buildpulse.log is present and contains expected content
		logdata, err := os.ReadFile(filepath.Join(unzipDir, "buildpulse.log"))
		require.NoError(t, err)
		assert.Contains(t, string(logdata), "Gathering metadata to describe the build")
	})

	t.Run("bundle with tags", func(t *testing.T) {
		envs := map[string]string{
			"GITHUB_ACTIONS": "true",
			"GITHUB_SHA":     "aaaaaaaaaaaaaaaaaaaabbbbbbbbbbbbbbbbbbbb",
		}

		log := logger.New()
		s := &Submit{
			logger:         log,
			version:        &metadata.Version{Number: "v1.2.3"},
			commitResolver: metadata.NewStaticCommitResolver(&metadata.Commit{TreeSHA: "ccccccccccccccccccccdddddddddddddddddddd"}, log),
			envs:           envs,
			paths:          []string{"testdata/example-reports-dir/example-1.xml"},
			bucket:         "buildpulse-uploads",
			accountID:      42,
			repositoryID:   8675309,
			tagsString:     "tag1 tag2",
		}

		path, err := s.bundle()
		require.NoError(t, err)

		unzipDir := t.TempDir()
		err = archiver.Unarchive(path, unzipDir)
		require.NoError(t, err)

		// Verify buildpulse.yml is present and contains expected content
		yaml, err := os.ReadFile(filepath.Join(unzipDir, "buildpulse.yml"))
		require.NoError(t, err)

		assert.Contains(t, string(yaml), "- tag1")
		assert.Contains(t, string(yaml), "- tag2")
	})
}

func Test_upload(t *testing.T) {
	tests := []struct {
		name            string
		fixture         string
		bucket          string
		accountID       uint64
		accessKeyID     string
		secretAccessKey string
		err             string
	}{
		{
			name:            "success",
			fixture:         "testdata/s3-success",
			bucket:          "buildpulse-uploads",
			accountID:       42,
			accessKeyID:     accessKeyID,
			secretAccessKey: secretAccessKey,
			err:             "",
		},
		{
			name:            "bad access key ID",
			fixture:         "testdata/s3-bad-access-key-id",
			bucket:          "buildpulse-uploads",
			accountID:       42,
			accessKeyID:     "some-bogus-access-key-id",
			secretAccessKey: secretAccessKey,
			err:             "InvalidAccessKeyId",
		},
		{
			name:            "bad secret access key",
			fixture:         "testdata/s3-bad-secret-access-key",
			bucket:          "buildpulse-uploads",
			accountID:       42,
			accessKeyID:     accessKeyID,
			secretAccessKey: "some-bogus-secret-access-key",
			err:             "SignatureDoesNotMatch",
		},
		{
			name:            "credentials not authorized for account ID",
			fixture:         "testdata/s3-unauthorized-object-prefix",
			bucket:          "buildpulse-uploads",
			accountID:       1,
			accessKeyID:     accessKeyID,
			secretAccessKey: secretAccessKey,
			err:             "AccessDenied",
		},
		{
			name:            "bad bucket",
			fixture:         "testdata/s3-bad-bucket",
			bucket:          "some-bogus-bucket",
			accountID:       42,
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
				logger:       logger.New(),
				bucket:       tt.bucket,
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
				assert.Equal(t, "42/8675309/buildpulse-00000000-0000-0000-0000-000000000000.gz", key)
			} else {
				assert.Contains(t, err.Error(), tt.err)
			}
		})
	}
}

func Test_toGz(t *testing.T) {
	path, err := toGz("testdata/example-reports-dir/example.txt")
	require.NoError(t, err)
	assert.Regexp(t, `\.gz$`, path)

	dir := t.TempDir()
	unzippedPath := filepath.Join(dir, "example.txt")
	err = archiver.DecompressFile(path, unzippedPath)
	require.NoError(t, err)

	// === Verify original content matches resulting content
	assertEqualContent(t, "testdata/example-reports-dir/example.txt", unzippedPath)
}

func Test_xmlPathsFromDir(t *testing.T) {
	tests := []struct {
		name string
		path string
		want []string
	}{
		{
			name: "DirectoryWithFilesAtRootAndInSubDirectories",
			path: "testdata/example-reports-dir",
			want: []string{
				"testdata/example-reports-dir/coverage/report.xml",
				"testdata/example-reports-dir/coverage/report-2.xml",
				"testdata/example-reports-dir/example-1.xml",
				"testdata/example-reports-dir/example-2.XML",
				"testdata/example-reports-dir/dir-with-xml-files/browserstack/example-1.xml",
				"testdata/example-reports-dir/dir-with-xml-files/browserstack/example-2.xml",
				"testdata/example-reports-dir/dir-with-xml-files/browsertest/example-3.xml",
			},
		},
		{
			name: "DirectoryWithoutXMLFiles",
			path: "testdata/example-reports-dir/dir-without-xml-files",
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reportPaths, err := xmlPathsFromDir(tt.path)
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.want, reportPaths)
		})
	}
}

func Test_xmlPathsFromGlob(t *testing.T) {
	tests := []struct {
		name string
		path string
		want []string
	}{
		{
			name: "ExactPathToSingleFile",
			path: "testdata/example-reports-dir/example-1.xml",
			want: []string{
				"testdata/example-reports-dir/example-1.xml",
			},
		},
		{
			name: "PathMatchingFilesByWildcard",
			path: "testdata/example-reports-dir/example*",
			want: []string{
				"testdata/example-reports-dir/example-1.xml",
				"testdata/example-reports-dir/example-2.XML",
			},
		},
		{
			name: "PathMatchingDirectoriesAndFilesByWildcard",
			path: "testdata/example-reports-dir/dir-with-xml-files/*/*.xml",
			want: []string{
				"testdata/example-reports-dir/dir-with-xml-files/browserstack/example-1.xml",
				"testdata/example-reports-dir/dir-with-xml-files/browserstack/example-2.xml",
				"testdata/example-reports-dir/dir-with-xml-files/browsertest/example-3.xml",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reportPaths, err := xmlPathsFromGlob(tt.path)
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.want, reportPaths)
		})
	}
}

// assertEqualContent asserts that two files have the same content.
func assertEqualContent(t *testing.T, expected string, actual string) {
	expectedBytes, err := os.ReadFile(expected)
	require.NoError(t, err)

	actualBytes, err := os.ReadFile(actual)
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
	r.Body = io.NopCloser(&b)
	return cassette.DefaultMatcher(r, i) && (b.String() == i.Body)
}

var _ metadata.CommitResolver = (*stubCommitResolver)(nil)

type stubCommitResolver struct {
	source string
}

func (s *stubCommitResolver) Lookup(sha string) (*metadata.Commit, error) {
	return &metadata.Commit{}, nil
}

func (s *stubCommitResolver) Source() string {
	return s.source
}

var _ CommitResolverFactory = (*stubCommitResolverFactory)(nil)

type stubCommitResolverFactory struct{}

func (s *stubCommitResolverFactory) NewFromRepository(path string) (metadata.CommitResolver, error) {
	return &stubCommitResolver{source: "Repository"}, nil
}

func (s *stubCommitResolverFactory) NewFromStaticValue(commit *metadata.Commit) metadata.CommitResolver {
	return &stubCommitResolver{source: "Static"}
}
