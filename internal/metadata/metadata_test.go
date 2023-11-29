package metadata

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/buildpulse/test-reporter/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetadata(t *testing.T) {
	tests := []struct {
		name    string
		envs    map[string]string
		tags    []string
		fixture string
	}{
		{
			name: "Buildkite",
			envs: map[string]string{
				"BUILDKITE_BRANCH":            "some-branch",
				"BUILDKITE_BUILD_ID":          "00000000-0000-0000-0000-000000000000",
				"BUILDKITE_BUILD_NUMBER":      "42",
				"BUILDKITE_BUILD_URL":         "https://buildkite.com/some-org/some-project/builds/8675309",
				"BUILDKITE_COMMIT":            "1f192ff735f887dd7a25229b2ece0422d17931f5",
				"BUILDKITE_JOB_ID":            "11111111-1111-1111-1111-111111111111",
				"BUILDKITE_LABEL":             ":test_tube: Run tests",
				"BUILDKITE_ORGANIZATION_SLUG": "some-org",
				"BUILDKITE_PIPELINE_ID":       "22222222-2222-2222-2222-222222222222",
				"BUILDKITE_PIPELINE_SLUG":     "some-pipeline",
				"BUILDKITE_PROJECT_SLUG":      "some-org/some-project",
				"BUILDKITE_PULL_REQUEST":      "false",
				"BUILDKITE_REPO":              "git@github.com:some-owner/some-repo.git",
				"BUILDKITE_RETRY_COUNT":       "2",
				"BUILDKITE":                   "true",
			},
			fixture: "./testdata/buildkite.yml",
		},
		{
			name: "CircleCI",
			envs: map[string]string{
				"CIRCLECI":                "true",
				"CIRCLE_BRANCH":           "some-branch",
				"CIRCLE_BUILD_NUM":        "1",
				"CIRCLE_BUILD_URL":        "https://circleci.com/gh/some-owner/some-repo/8675309",
				"CIRCLE_SHA1":             "1f192ff735f887dd7a25229b2ece0422d17931f5",
				"CIRCLE_JOB":              "some-job",
				"CIRCLE_PROJECT_REPONAME": "some-repo",
				"CIRCLE_PROJECT_USERNAME": "some-owner",
				"CIRCLE_REPOSITORY_URL":   "git@github.com:some-owner/some-repo.git",
				"CIRCLE_WORKFLOW_ID":      "00000000-0000-0000-0000-000000000000",
				"CIRCLE_USERNAME":         "some-committer",
			},
			fixture: "./testdata/circle.yml",
		},
		{
			name: "GitHubActions",
			envs: map[string]string{
				"GITHUB_ACTIONS":     "true",
				"GITHUB_ACTOR":       "some-user",
				"GITHUB_BASE_REF":    "refs/heads/main",
				"GITHUB_EVENT_NAME":  "push",
				"GITHUB_HEAD_REF":    "refs/heads/some-feature",
				"GITHUB_REF":         "refs/heads/some-feature",
				"GITHUB_REPOSITORY":  "some-owner/some-repo",
				"GITHUB_RUN_ATTEMPT": "1",
				"GITHUB_RUN_ID":      "8675309",
				"GITHUB_RUN_NUMBER":  "42",
				"GITHUB_SERVER_URL":  "https://github.com",
				"GITHUB_SHA":         "1f192ff735f887dd7a25229b2ece0422d17931f5",
				"GITHUB_WORKFLOW":    "build",
			},
			fixture: "./testdata/github.yml",
		},
		{
			name: "Jenkins",
			envs: map[string]string{
				"BUILD_URL":       "https://some-jenkins-server.com/job/some-project/8675309",
				"EXECUTOR_NUMBER": "42",
				"GIT_BRANCH":      "origin/some-branch",
				"GIT_COMMIT":      "1f192ff735f887dd7a25229b2ece0422d17931f5",
				"GIT_URL":         "https://github.com/some-owner/some-repo.git",
				"JENKINS_HOME":    "/var/lib/jenkins",
				"JOB_NAME":        "some-project",
				"JOB_URL":         "https://some-jenkins-server.com/job/some-project/",
				"NODE_NAME":       "master",
				"WORKSPACE":       "/var/lib/jenkins/workspace/some-project",
			},
			fixture: "./testdata/jenkins.yml",
		},
		{
			name: "Semaphore",
			envs: map[string]string{
				"SEMAPHORE": "true",
				"SEMAPHORE_AGENT_MACHINE_ENVIRONMENT_TYPE": "container",
				"SEMAPHORE_AGENT_MACHINE_OS_IMAGE":         "ubuntu1804",
				"SEMAPHORE_AGENT_MACHINE_TYPE":             "e1-standard-4",
				"SEMAPHORE_GIT_BRANCH":                     "some-branch",
				"SEMAPHORE_GIT_COMMIT_RANGE":               "39ae7373c5a46bfda3ff2c5e2f58b789eab9bfcd...1f192ff735f887dd7a25229b2ece0422d17931f5",
				"SEMAPHORE_GIT_DIR":                        "some-dir",
				"SEMAPHORE_GIT_REF_TYPE":                   "branch",
				"SEMAPHORE_GIT_REF":                        "refs/heads/main",
				"SEMAPHORE_GIT_REPO_SLUG":                  "some-owner/some-repo",
				"SEMAPHORE_GIT_SHA":                        "1f192ff735f887dd7a25229b2ece0422d17931f5",
				"SEMAPHORE_GIT_URL":                        "git@github.com:some-owner/some-repo.git",
				"SEMAPHORE_JOB_ID":                         "00000000-0000-0000-0000-000000000000",
				"SEMAPHORE_JOB_NAME":                       "Run tests",
				"SEMAPHORE_JOB_RESULT":                     "passed",
				"SEMAPHORE_ORGANIZATION_URL":               "https://some-owner.semaphoreci.com",
				"SEMAPHORE_PROJECT_ID":                     "11111111-1111-1111-1111-111111111111",
				"SEMAPHORE_PROJECT_NAME":                   "some-repo",
				"SEMAPHORE_WORKFLOW_ID":                    "22222222-2222-2222-2222-222222222222",
				"SEMAPHORE_WORKFLOW_NUMBER":                "42",
			},
			fixture: "./testdata/semaphore.yml",
		},
		{
			name: "Travis",
			envs: map[string]string{
				"TRAVIS_BRANCH":              "some-branch",
				"TRAVIS_BUILD_DIR":           "/home/travis/build/some-owner/some-repo",
				"TRAVIS_BUILD_ID":            "1111111",
				"TRAVIS_BUILD_NUMBER":        "42",
				"TRAVIS_BUILD_WEB_URL":       "https://travis-ci.org/some-owner/some-repo-guacamole/builds/1111111",
				"TRAVIS_COMMIT_RANGE":        "39ae7373c5a4...1f192ff735f8",
				"TRAVIS_COMMIT":              "1f192ff735f887dd7a25229b2ece0422d17931f5",
				"TRAVIS_CPU_ARCH":            "amd64",
				"TRAVIS_DIST":                "xenial",
				"TRAVIS_EVENT_TYPE":          "push",
				"TRAVIS_JOB_ID":              "8675309",
				"TRAVIS_JOB_NAME":            "",
				"TRAVIS_JOB_NUMBER":          "42.1",
				"TRAVIS_JOB_WEB_URL":         "https://travis-ci.org/some-owner/some-repo/jobs/8675309",
				"TRAVIS_OS_NAME":             "linux",
				"TRAVIS_PULL_REQUEST_BRANCH": "",
				"TRAVIS_PULL_REQUEST_SHA":    "",
				"TRAVIS_PULL_REQUEST_SLUG":   "",
				"TRAVIS_PULL_REQUEST":        "false",
				"TRAVIS_REPO_SLUG":           "some-owner/some-repo",
				"TRAVIS_SUDO":                "true",
				"TRAVIS_TAG":                 "",
				"TRAVIS_TEST_RESULT":         "0",
				"TRAVIS":                     "true",
			},
			fixture: "./testdata/travis.yml",
		},
		{
			name: "webapp.io",
			envs: map[string]string{
				"GIT_BRANCH":        "some-branch",
				"GIT_COMMIT":        "1f192ff735f887dd7a25229b2ece0422d17931f5",
				"JOB_ID":            "42",
				"ORGANIZATION_NAME": "some-org",
				"REPOSITORY_NAME":   "some-repo",
				"REPOSITORY_OWNER":  "some-owner",
				"RETRY_INDEX":       "1",
				"RUNNER_ID":         "main-layerfile",
				"WEBAPPIO":          "true",
			},
			fixture: "./testdata/webapp.io.yml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := func() time.Time {
				return time.Date(2020, 7, 11, 1, 2, 3, 0, time.UTC)
			}

			expected, err := os.ReadFile(tt.fixture)
			require.NoError(t, err)

			authoredAt, err := time.Parse(time.RFC3339, "2020-07-09T04:05:06-05:00")
			require.NoError(t, err)

			committedAt, err := time.Parse(time.RFC3339, "2020-07-10T07:08:09+13:00")
			require.NoError(t, err)

			commitResolver := NewStaticCommitResolver(
				&Commit{
					AuthoredAt:     authoredAt,
					AuthorEmail:    "some-author@example.com",
					AuthorName:     "Some Author",
					CommittedAt:    committedAt,
					CommitterEmail: "some-committer@example.com",
					CommitterName:  "Some Committer",
					Message:        "Some message",
					TreeSHA:        "0da9df599c02da5e7f5058b7108dcd5e1929a0fe",
				},
				logger.New(),
			)

			version := &Version{Number: "v1.2.3", GoOS: "linux"}
			meta, err := NewMetadata(version, tt.envs, tt.tags, "", commitResolver, now, logger.New())
			assert.NoError(t, err)

			yaml, err := meta.MarshalYAML()
			assert.NoError(t, err)
			assert.Equal(t, string(expected), string(yaml))
		})
	}
}

func TestNewMetadata_unsupportedProvider(t *testing.T) {
	_, err := NewMetadata(&Version{}, map[string]string{}, []string{}, "", newCommitResolverStub(), time.Now, logger.New())
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "env: environment variable \"GIT_BRANCH\" should not be empty; environment variable \"GIT_COMMIT\" should not be empty; environment variable \"BUILD_URL\" should not be empty; environment variable \"ORGANIZATION_NAME\" should not be empty; environment variable \"REPOSITORY_NAME\" should not be empty")
	}
}

func TestNewMetadata_customCheckName(t *testing.T) {
	tests := []struct {
		name          string
		envs          map[string]string
		expectedCheck string
	}{
		{
			name: "with custom check name present",
			envs: map[string]string{
				"BUILDPULSE_CHECK_NAME": "some-custom-check-name",
				"GITHUB_ACTIONS":        "true",
			},
			expectedCheck: "some-custom-check-name",
		},
		{
			name: "with custom check name present but empty",
			envs: map[string]string{
				"BUILDPULSE_CHECK_NAME": "",
				"GITHUB_ACTIONS":        "true",
			},
			expectedCheck: "github-actions",
		},
		{
			name: "without custom check name",
			envs: map[string]string{
				"GITHUB_ACTIONS": "true",
			},
			expectedCheck: "github-actions",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta, err := NewMetadata(&Version{}, tt.envs, []string{}, "", newCommitResolverStub(), time.Now, logger.New())
			assert.NoError(t, err)

			yaml, err := meta.MarshalYAML()
			assert.NoError(t, err)
			assert.Regexp(t, fmt.Sprintf(":check: %s", tt.expectedCheck), string(yaml))
		})
	}
}

func newCommitResolverStub() CommitResolver {
	return NewStaticCommitResolver(&Commit{}, logger.New())
}

func TestNewMetadata_appliesTags(t *testing.T) {
	tests := []struct {
		name    string
		envs    map[string]string
		tags    []string
		fixture string
	}{
		{
			name: "GitHubActions",
			envs: map[string]string{
				"GITHUB_ACTIONS":     "true",
				"GITHUB_ACTOR":       "some-user",
				"GITHUB_BASE_REF":    "refs/heads/main",
				"GITHUB_EVENT_NAME":  "push",
				"GITHUB_HEAD_REF":    "refs/heads/some-feature",
				"GITHUB_REF":         "refs/heads/some-feature",
				"GITHUB_REPOSITORY":  "some-owner/some-repo",
				"GITHUB_RUN_ATTEMPT": "1",
				"GITHUB_RUN_ID":      "8675309",
				"GITHUB_RUN_NUMBER":  "42",
				"GITHUB_SERVER_URL":  "https://github.com",
				"GITHUB_SHA":         "1f192ff735f887dd7a25229b2ece0422d17931f5",
				"GITHUB_WORKFLOW":    "build",
			},
			fixture: "./testdata/github_tags.yml",
			tags:    []string{"tag1", "tag2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := func() time.Time {
				return time.Date(2020, 7, 11, 1, 2, 3, 0, time.UTC)
			}

			expected, err := os.ReadFile(tt.fixture)
			require.NoError(t, err)

			authoredAt, err := time.Parse(time.RFC3339, "2020-07-09T04:05:06-05:00")
			require.NoError(t, err)

			committedAt, err := time.Parse(time.RFC3339, "2020-07-10T07:08:09+13:00")
			require.NoError(t, err)

			commitResolver := NewStaticCommitResolver(
				&Commit{
					AuthoredAt:     authoredAt,
					AuthorEmail:    "some-author@example.com",
					AuthorName:     "Some Author",
					CommittedAt:    committedAt,
					CommitterEmail: "some-committer@example.com",
					CommitterName:  "Some Committer",
					Message:        "Some message",
					TreeSHA:        "0da9df599c02da5e7f5058b7108dcd5e1929a0fe",
				},
				logger.New(),
			)

			version := &Version{Number: "v1.2.3", GoOS: "linux"}
			meta, err := NewMetadata(version, tt.envs, tt.tags, "", commitResolver, now, logger.New())
			assert.NoError(t, err)

			yaml, err := meta.MarshalYAML()
			assert.NoError(t, err)
			assert.Equal(t, string(expected), string(yaml))
		})
	}
}

func TestNewMetadata_appliesQuotaID(t *testing.T) {
	tests := []struct {
		name    string
		envs    map[string]string
		quotaID string
		fixture string
	}{
		{
			name: "GitHubActions",
			envs: map[string]string{
				"GITHUB_ACTIONS":     "true",
				"GITHUB_ACTOR":       "some-user",
				"GITHUB_BASE_REF":    "refs/heads/main",
				"GITHUB_EVENT_NAME":  "push",
				"GITHUB_HEAD_REF":    "refs/heads/some-feature",
				"GITHUB_REF":         "refs/heads/some-feature",
				"GITHUB_REPOSITORY":  "some-owner/some-repo",
				"GITHUB_RUN_ATTEMPT": "1",
				"GITHUB_RUN_ID":      "8675309",
				"GITHUB_RUN_NUMBER":  "42",
				"GITHUB_SERVER_URL":  "https://github.com",
				"GITHUB_SHA":         "1f192ff735f887dd7a25229b2ece0422d17931f5",
				"GITHUB_WORKFLOW":    "build",
			},
			fixture: "./testdata/github_quota.yml",
			quotaID: "quota1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := func() time.Time {
				return time.Date(2020, 7, 11, 1, 2, 3, 0, time.UTC)
			}

			expected, err := os.ReadFile(tt.fixture)
			require.NoError(t, err)

			authoredAt, err := time.Parse(time.RFC3339, "2020-07-09T04:05:06-05:00")
			require.NoError(t, err)

			committedAt, err := time.Parse(time.RFC3339, "2020-07-10T07:08:09+13:00")
			require.NoError(t, err)

			commitResolver := NewStaticCommitResolver(
				&Commit{
					AuthoredAt:     authoredAt,
					AuthorEmail:    "some-author@example.com",
					AuthorName:     "Some Author",
					CommittedAt:    committedAt,
					CommitterEmail: "some-committer@example.com",
					CommitterName:  "Some Committer",
					Message:        "Some message",
					TreeSHA:        "0da9df599c02da5e7f5058b7108dcd5e1929a0fe",
				},
				logger.New(),
			)

			version := &Version{Number: "v1.2.3", GoOS: "linux"}
			meta, err := NewMetadata(version, tt.envs, []string{}, tt.quotaID, commitResolver, now, logger.New())
			assert.NoError(t, err)

			yaml, err := meta.MarshalYAML()
			assert.NoError(t, err)
			assert.Equal(t, string(expected), string(yaml))
		})
	}
}
