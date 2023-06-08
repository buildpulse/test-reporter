package metadata

import (
	"testing"

	"github.com/buildpulse/test-reporter/internal/logger"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func Test_azurePipelinesMetadata_Init_extraFields(t *testing.T) {
	tests := []struct {
		name          string
		envs          map[string]string
		expectedLines []string
	}{
		{
			name: "with github source",
			envs: map[string]string{
				"BUILD_SOURCEVERSION":                "00000000-0000-0000-0000-000000000000",
				"BUILD_SOURCEBRANCHNAME":             "some-branch-name",
				"BUILD_BUILDURI":                     "https://github.com/some-forker/some-repo",
				"BUILD_REPOSITORY_NAME":              "org/some-repo",
				"SYSTEM_TEAMFOUNDATIONCOLLECTIONURI": "org",
			},
			expectedLines: []string{
				":build_sourceversion: 00000000-0000-0000-0000-000000000000",
				":build_sourcebranchname: some-branch-name",
				":build_builduri: https://github.com/some-forker/some-repo",
				":build_repository_name: org/some-repo",
				":system_teamfoundationcollectionuri: org",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := azurePipelinesMetadata{}
			err := meta.Init(tt.envs, logger.New())
			assert.NoError(t, err)

			yaml, err := yaml.Marshal(meta)
			assert.NoError(t, err)
			for _, line := range tt.expectedLines {
				assert.Regexp(t, line, string(yaml))
			}
		})
	}
}

func Test_bitbucketMetadata_Init_extraFields(t *testing.T) {
	tests := []struct {
		name          string
		envs          map[string]string
		expectedLines []string
	}{
		{
			name: "with pull request",
			envs: map[string]string{
				"BITBUCKET_COMMIT":          "00000000-0000-0000-0000-000000000000",
				"BITBUCKET_BRANCH":          "some-branch-name",
				"BITBUCKET_GIT_HTTP_ORIGIN": "https://github.com/some-forker/some-repo",
				"BITBUCKET_WORKSPACE":       "org",
				"BITBUCKET_BUILD_NUMBER":    "99",
				"BITBUCKET_REPO_SLUG":       "some-repo",
			},
			expectedLines: []string{
				":bitbucket_commit: 00000000-0000-0000-0000-000000000000",
				":bitbucket_branch: some-branch-name",
				":bitbucket_git_http_origin: https://github.com/some-forker/some-repo",
				":bitbucket_workspace: org",
				":bitbucket_build_number: 99",
				":bitbucket_repo_slug: some-repo",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := bitbucketMetadata{}
			err := meta.Init(tt.envs, logger.New())
			assert.NoError(t, err)

			yaml, err := yaml.Marshal(meta)
			assert.NoError(t, err)
			for _, line := range tt.expectedLines {
				assert.Regexp(t, line, string(yaml))
			}
		})
	}
}

func Test_awsCodeBuildMetadata_Init_extraFields(t *testing.T) {
	tests := []struct {
		name          string
		envs          map[string]string
		expectedLines []string
	}{
		{
			name: "with pull request",
			envs: map[string]string{
				"CODEBUILD_SOURCE_VERSION":          "pr/99",
				"CODEBUILD_PUBLIC_BUILD_URL":        "https://github.com/some-forker/some-repo",
				"CODEBUILD_RESOLVED_SOURCE_VERSION": "00000000-0000-0000-0000-000000000000",
			},
			expectedLines: []string{
				":codebuild_source_version: pr/99",
				":codebuild_public_build_url: https://github.com/some-forker/some-repo",
				":codebuild_resolved_source_version: 00000000-0000-0000-0000-000000000000",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := awsCodeBuildMetadata{}
			err := meta.Init(tt.envs, logger.New())
			assert.NoError(t, err)

			yaml, err := yaml.Marshal(meta)
			assert.NoError(t, err)
			for _, line := range tt.expectedLines {
				assert.Regexp(t, line, string(yaml))
			}
		})
	}
}

func Test_buildkiteMetadata_Init_extraFields(t *testing.T) {
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
			err := meta.Init(tt.envs, logger.New())
			assert.NoError(t, err)

			yaml, err := yaml.Marshal(meta)
			assert.NoError(t, err)
			for _, line := range tt.expectedLines {
				assert.Regexp(t, line, string(yaml))
			}
		})
	}
}

func Test_circleMetadata_Init_extraFields(t *testing.T) {
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
			err := meta.Init(tt.envs, logger.New())
			assert.NoError(t, err)

			yaml, err := yaml.Marshal(meta)
			assert.NoError(t, err)
			for _, line := range tt.expectedLines {
				assert.Regexp(t, line, string(yaml))
			}
		})
	}
}

func Test_githubMetadata_Init_repoURL(t *testing.T) {
	tests := []struct {
		name string
		envs map[string]string
		want string
	}{
		{
			name: "when GITHUB_SERVER_URL is present",
			envs: map[string]string{
				"GITHUB_REPOSITORY": "some-owner/some-repo",
				"GITHUB_SERVER_URL": "https://github.com",
			},
			want: "https://github.com/some-owner/some-repo",
		},
		{
			name: "when GITHUB_SERVER_URL is blank",
			envs: map[string]string{
				"GITHUB_REPOSITORY": "some-owner/some-repo",
				"GITHUB_SERVER_URL": "",
			},
			want: "https://github.com/some-owner/some-repo",
		},
		{
			name: "when GITHUB_SERVER_URL is missing",
			envs: map[string]string{
				"GITHUB_REPOSITORY": "some-owner/some-repo",
			},
			want: "https://github.com/some-owner/some-repo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := githubMetadata{}
			err := meta.Init(tt.envs, logger.New())
			assert.NoError(t, err)

			assert.Equal(t, tt.want, meta.GithubRepoURL)
		})
	}
}
func Test_githubMetadata_Init_branchIdentification(t *testing.T) {
	tests := []struct {
		name string
		envs map[string]string
		want string
	}{
		{
			name: "push event for branch",
			envs: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF":        "refs/heads/some-branch",
			},
			want: "some-branch",
		},
		{
			name: "push event for tag",
			envs: map[string]string{
				"GITHUB_BASE_REF":   "",
				"GITHUB_EVENT_NAME": "push",
				"GITHUB_HEAD_REF":   "",
				"GITHUB_REF":        "refs/tags/v0.1.0",
			},
			want: "",
		},
		{
			name: "pull request event for branch from source repo ",
			envs: map[string]string{
				"GITHUB_BASE_REF":   "main",
				"GITHUB_EVENT_NAME": "pull_request",
				"GITHUB_HEAD_REF":   "some-branch",
				"GITHUB_REF":        "refs/pull/2/merge",
			},
			want: "some-branch",
		},
		{
			name: "pull request event for branch from fork",
			envs: map[string]string{
				"GITHUB_BASE_REF":   "main",
				"GITHUB_EVENT_NAME": "pull_request",
				"GITHUB_HEAD_REF":   "some-branch",
				"GITHUB_REF":        "refs/pull/3/merge",
			},
			want: "some-branch",
		},
		{
			name: "neither a branch nor a tag",
			envs: map[string]string{}, // The GITHUB_REF env var is not present in this scenario
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := githubMetadata{}
			err := meta.Init(tt.envs, logger.New())
			assert.NoError(t, err)

			assert.Equal(t, tt.want, meta.Branch())
		})
	}
}

func Test_travisMetadata_Init_extraFields(t *testing.T) {
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
			err := meta.Init(tt.envs, logger.New())
			assert.NoError(t, err)

			yaml, err := yaml.Marshal(meta)
			assert.NoError(t, err)
			for _, line := range tt.expectedLines {
				assert.Regexp(t, line, string(yaml))
			}
		})
	}
}

func Test_webappioMetadata_Init_extraFields(t *testing.T) {
	tests := []struct {
		name          string
		envs          map[string]string
		expectedLines []string
	}{
		{
			name: "with pull request",
			envs: map[string]string{
				"PULL_REQUEST_URL": "https://github.com/some-owner/some-repo/pull/1",
			},
			expectedLines: []string{
				":pull_request_url: https://github.com/some-owner/some-repo/pull/1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := webappioMetadata{}
			err := meta.Init(tt.envs, logger.New())
			assert.NoError(t, err)

			yaml, err := yaml.Marshal(meta)
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
