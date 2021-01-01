package metadata

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetadata(t *testing.T) {
	tests := []struct {
		name    string
		envs    map[string]string
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
				"GITHUB_ACTIONS":    "true",
				"GITHUB_ACTOR":      "some-user",
				"GITHUB_BASE_REF":   "refs/heads/main",
				"GITHUB_HEAD_REF":   "refs/heads/some-feature",
				"GITHUB_REF":        "refs/heads/some-feature",
				"GITHUB_REPOSITORY": "some-owner/some-repo",
				"GITHUB_RUN_ID":     "8675309",
				"GITHUB_RUN_NUMBER": "42",
				"GITHUB_SERVER_URL": "https://github.com",
				"GITHUB_SHA":        "1f192ff735f887dd7a25229b2ece0422d17931f5",
				"GITHUB_WORKFLOW":   "build",
			},
			fixture: "./testdata/github.yml",
		},
		{
			name: "Jenkins",
			envs: map[string]string{
				"BUILD_URL":    "https://some-jenkins-server.com/job/some-project/8675309",
				"GIT_BRANCH":   "origin/some-branch",
				"GIT_COMMIT":   "1f192ff735f887dd7a25229b2ece0422d17931f5",
				"GIT_URL":      "https://github.com/some-owner/some-repo.git",
				"JENKINS_HOME": "/var/lib/jenkins",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := func() time.Time {
				return time.Date(2020, 7, 11, 1, 2, 3, 0, time.UTC)
			}

			expected, err := ioutil.ReadFile(tt.fixture)
			require.NoError(t, err)

			commitResolverDouble := CommitResolverFunc(
				func(sha string) (*Commit, error) {
					return &Commit{
						SHA:     sha,
						TreeSHA: "0da9df599c02da5e7f5058b7108dcd5e1929a0fe",
					}, nil
				})

			version := &Version{Number: "v1.2.3", GoOS: "linux"}
			meta, err := NewMetadata(version, tt.envs, commitResolverDouble, now)
			assert.NoError(t, err)

			yaml, err := meta.MarshalYAML()
			assert.NoError(t, err)
			assert.Equal(t, string(expected), string(yaml))
		})
	}
}

func TestNewMetadata_unsupportedProvider(t *testing.T) {
	_, err := NewMetadata(&Version{}, map[string]string{}, newCommitResolverStub(), time.Now)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "unrecognized environment")
	}
}

func TestNewMetadata_customCheckName(t *testing.T) {
	tests := []struct {
		name             string
		envs             map[string]string
		expectedProvider string
		expectedCheck    string
	}{
		{
			name: "Buildkite",
			envs: map[string]string{
				"BUILDKITE":             "true",
				"BUILDPULSE_CHECK_NAME": "some-custom-check-name",
				"BUILDKITE_REPO":        "git@github.com:x/y.git",
			},
			expectedProvider: "buildkite",
			expectedCheck:    "some-custom-check-name",
		},
		{
			name: "Circle",
			envs: map[string]string{
				"CIRCLECI":              "true",
				"BUILDPULSE_CHECK_NAME": "some-custom-check-name",
			},
			expectedProvider: "circleci",
			expectedCheck:    "some-custom-check-name",
		},
		{
			name: "GitHubActions",
			envs: map[string]string{
				"GITHUB_ACTIONS":        "true",
				"BUILDPULSE_CHECK_NAME": "some-custom-check-name",
			},
			expectedProvider: "github-actions",
			expectedCheck:    "some-custom-check-name",
		},
		{
			name: "Jenkins",
			envs: map[string]string{
				"JENKINS_HOME":          "/var/lib/jenkins",
				"BUILDPULSE_CHECK_NAME": "some-custom-check-name",
				"BUILD_URL":             "https://some-jenkins-server.com/job/some-project/8675309",
				"GIT_URL":               "https://github.com/some-owner/some-repo.git",
			},
			expectedProvider: "jenkins",
			expectedCheck:    "some-custom-check-name",
		},
		{
			name: "Semaphore",
			envs: map[string]string{
				"SEMAPHORE":             "true",
				"BUILDPULSE_CHECK_NAME": "some-custom-check-name",
			},
			expectedProvider: "semaphore",
			expectedCheck:    "some-custom-check-name",
		},
		{
			name: "Travis",
			envs: map[string]string{
				"TRAVIS":                "true",
				"BUILDPULSE_CHECK_NAME": "some-custom-check-name",
			},
			expectedProvider: "travis-ci",
			expectedCheck:    "some-custom-check-name",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta, err := NewMetadata(&Version{}, tt.envs, newCommitResolverStub(), time.Now)
			assert.NoError(t, err)

			yaml, err := meta.MarshalYAML()
			assert.NoError(t, err)
			assert.Regexp(t, fmt.Sprintf(":ci_provider: %s", tt.expectedProvider), string(yaml))
			assert.Regexp(t, fmt.Sprintf(":check: %s", tt.expectedCheck), string(yaml))
		})
	}
}

func Test_buildkiteMetadata_initEnvData_extraFields(t *testing.T) {
	tests := []struct {
		name          string
		envs          map[string]string
		expectedLines []string
	}{
		{
			name: "when rebuilt",
			envs: map[string]string{
				"BUILDKITE_REBUILT_FROM_BUILD_ID":     "00000000-0000-0000-0000-000000000000",
				"BUILDKITE_REBUILT_FROM_BUILD_NUMBER": "42",
				"BUILDKITE_REPO":                      "git@github.com:x/y.git",
			},
			expectedLines: []string{
				":buildkite_rebuilt_from_build_id: 00000000-0000-0000-0000-000000000000",
				":buildkite_rebuilt_from_build_number: 42",
			},
		},
		{
			name: "with pull request",
			envs: map[string]string{
				"BUILDKITE_PULL_REQUEST_BASE_BRANCH": "some-base-branch",
				"BUILDKITE_PULL_REQUEST_REPO":        "git://github.com/some-forker/some-repo.git",
				"BUILDKITE_PULL_REQUEST":             "99",
				"BUILDKITE_REPO":                     "git@github.com:x/y.git",
			},
			expectedLines: []string{
				":buildkite_pull_request_base_branch: some-base-branch",
				":buildkite_pull_request_repo: git://github.com/some-forker/some-repo.git",
				":buildkite_pull_request_number: 99",
			},
		},
		{
			name: "with tag",
			envs: map[string]string{
				"BUILDKITE_REPO": "git@github.com:x/y.git",
				"BUILDKITE_TAG":  "v0.1.0",
			},
			expectedLines: []string{":buildkite_tag: v0.1.0"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := buildkiteMetadata{}
			err := meta.initEnvData(tt.envs, newCommitResolverStub())
			assert.NoError(t, err)

			yaml, err := meta.MarshalYAML()
			assert.NoError(t, err)
			for _, line := range tt.expectedLines {
				assert.Regexp(t, line, string(yaml))
			}
		})
	}
}

func Test_circleMetadata_initEnvData_extraFields(t *testing.T) {
	tests := []struct {
		name          string
		envs          map[string]string
		expectedLines []string
	}{
		{
			name: "with pull request",
			envs: map[string]string{
				"CIRCLE_PR_NUMBER":    "42",
				"CIRCLE_PR_REPONAME":  "some-repo",
				"CIRCLE_PR_USERNAME":  "some-forker",
				"CIRCLE_PULL_REQUEST": "https://github.com/some-owner/some-repo/pull/42",
			},
			expectedLines: []string{
				":circle_pr_number: 42",
				":circle_pr_reponame: some-repo",
				":circle_pr_username: some-forker",
				":circle_pull_request: https://github.com/some-owner/some-repo/pull/42",
			},
		},
		{
			name: "with tag",
			envs: map[string]string{
				"CIRCLE_TAG": "v0.1.0",
			},
			expectedLines: []string{":circle_tag: v0.1.0"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := circleMetadata{}
			err := meta.initEnvData(tt.envs, newCommitResolverStub())
			assert.NoError(t, err)

			yaml, err := meta.MarshalYAML()
			assert.NoError(t, err)
			for _, line := range tt.expectedLines {
				assert.Regexp(t, line, string(yaml))
			}
		})
	}
}

func Test_githubMetadata_initEnvData_refTypes(t *testing.T) {
	tests := []struct {
		name string
		envs map[string]string
		yaml string
	}{
		{
			name: "branch",
			envs: map[string]string{
				"GITHUB_REF": "refs/heads/some-branch",
			},
			yaml: ":branch: some-branch",
		},
		{
			name: "tag",
			envs: map[string]string{
				"GITHUB_REF": "refs/tags/v0.1.0",
			},
			yaml: ":branch: \"\"\n",
		},
		{
			name: "neither a branch nor a tag",
			envs: map[string]string{}, // The GITHUB_REF env var is not present in this scenario
			yaml: ":branch: \"\"\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := githubMetadata{}
			err := meta.initEnvData(tt.envs, newCommitResolverStub())
			assert.NoError(t, err)

			yaml, err := meta.MarshalYAML()
			assert.NoError(t, err)
			assert.Contains(t, string(yaml), tt.yaml)
		})
	}
}

func Test_travisMetadata_initEnvData_extraFields(t *testing.T) {
	tests := []struct {
		name          string
		envs          map[string]string
		expectedLines []string
	}{
		{
			name: "with job name",
			envs: map[string]string{
				"TRAVIS_JOB_NAME": "some-job-name",
			},
			expectedLines: []string{":travis_job_name: some-job-name"},
		},
		{
			name: "with pull request",
			envs: map[string]string{
				"TRAVIS_PULL_REQUEST_BRANCH": "some-branch",
				"TRAVIS_PULL_REQUEST_SHA":    "eea22cb17a834f39961499af910ec96c82b035f4",
				"TRAVIS_PULL_REQUEST_SLUG":   "some-forker/some-repo",
				"TRAVIS_PULL_REQUEST":        "1",
			},
			expectedLines: []string{
				":travis_pull_request_branch: some-branch",
				":travis_pull_request_sha: eea22cb17a834f39961499af910ec96c82b035f4",
				":travis_pull_request_slug: some-forker/some-repo",
				":travis_pull_request_number: 1",
			},
		},
		{
			name: "with tag",
			envs: map[string]string{
				"TRAVIS_TAG": "v0.1.0",
			},
			expectedLines: []string{":travis_tag: v0.1.0"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := travisMetadata{}
			err := meta.initEnvData(tt.envs, newCommitResolverStub())
			assert.NoError(t, err)

			yaml, err := meta.MarshalYAML()
			assert.NoError(t, err)
			for _, line := range tt.expectedLines {
				assert.Regexp(t, line, string(yaml))
			}
		})
	}
}

func Test_nameWithOwnerFromGitURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		nwo  string
		err  bool
	}{
		{name: "https", url: "https://github.com/some-owner/some-repo.git", nwo: "some-owner/some-repo", err: false},
		{name: "ssh", url: "git@github.com:some-owner/some-repo.git", nwo: "some-owner/some-repo", err: false},
		{name: "malformed", url: "some-malformed-url", nwo: "", err: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nwo, err := nameWithOwnerFromGitURL(tt.url)

			if tt.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, nwo, tt.nwo)
			}
		})
	}
}

func newCommitResolverStub() CommitResolver {
	return CommitResolverFunc(
		func(sha string) (*Commit, error) {
			return &Commit{}, nil
		})
}
