package metadata

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
	"gopkg.in/yaml.v2"
)

// A Metadata instance provides metadata about a set of test results. It
// identifies the CI provider, the commit SHA, the time at which the tests were
// executed, etc.
type Metadata interface {
	MarshalYAML() (out []byte, err error)

	initEnvData(envs map[string]string, resolver CommitResolver) error
	initTimestamp(now func() time.Time)
	initVersionData(version *Version)
}

// AbstractMetadata provides the fields that are common across all Metadata
// instances, regardless of the specific CI provider.
type AbstractMetadata struct {
	AuthoredAt        time.Time `yaml:":authored_at"`
	AuthorEmail       string    `yaml:":author_email"`
	AuthorName        string    `yaml:":author_name"`
	Branch            string    `yaml:":branch"`
	BuildURL          string    `yaml:":build_url"`
	Check             string    `yaml:":check" env:"BUILDPULSE_CHECK_NAME"`
	CIProvider        string    `yaml:":ci_provider"`
	CommitMessage     string    `yaml:":commit_message"`
	CommitSHA         string    `yaml:":commit"`
	CommittedAt       time.Time `yaml:":committed_at"`
	CommitterEmail    string    `yaml:":committer_email"`
	CommitterName     string    `yaml:":committer_name"`
	RepoNameWithOwner string    `yaml:":repo_name_with_owner"`
	ReporterOS        string    `yaml:":reporter_os"`
	ReporterVersion   string    `yaml:":reporter_version"`
	Timestamp         time.Time `yaml:":timestamp"`
	TreeSHA           string    `yaml:":tree"`
}

func (a *AbstractMetadata) initCommitData(cr CommitResolver, sha string) error {
	c, err := cr.Lookup(sha)
	if err != nil {
		return err
	}

	a.AuthoredAt = c.AuthoredAt
	a.AuthorEmail = c.AuthorEmail
	a.AuthorName = c.AuthorName
	a.CommitMessage = c.Message
	a.CommitSHA = c.SHA
	a.CommittedAt = c.CommittedAt
	a.CommitterEmail = c.CommitterEmail
	a.CommitterName = c.CommitterName
	a.TreeSHA = c.TreeSHA

	return nil
}

func (a *AbstractMetadata) initTimestamp(now func() time.Time) {
	a.Timestamp = now()
}

func (a *AbstractMetadata) initVersionData(version *Version) {
	a.ReporterOS = version.GoOS
	a.ReporterVersion = version.Number
}

// NewMetadata creates a new Metadata instance from the given args.
func NewMetadata(version *Version, envs map[string]string, resolver CommitResolver, now func() time.Time) (Metadata, error) {
	var m Metadata

	switch {
	case envs["BUILDKITE"] == "true":
		m = &buildkiteMetadata{}
	case envs["CIRCLECI"] == "true":
		m = &circleMetadata{}
	case envs["GITHUB_ACTIONS"] == "true":
		m = &githubMetadata{}
	case envs["JENKINS_HOME"] != "":
		m = &jenkinsMetadata{}
	case envs["SEMAPHORE"] == "true":
		m = &semaphoreMetadata{}
	case envs["TRAVIS"] == "true":
		m = &travisMetadata{}
	default:
		return nil, fmt.Errorf("unrecognized environment: system does not appear to be a supported CI provider (Buildkite, CircleCI, GitHub Actions, Jenkins, Semaphore, or Travis CI)")
	}

	if err := m.initEnvData(envs, resolver); err != nil {
		return nil, err
	}
	m.initTimestamp(now)
	m.initVersionData(version)

	return m, nil
}

var _ Metadata = (*buildkiteMetadata)(nil)

type buildkiteMetadata struct {
	AbstractMetadata `yaml:",inline"`

	BuildkiteBranch                 string `env:"BUILDKITE_BRANCH" yaml:"-"`
	BuildkiteBuildID                string `env:"BUILDKITE_BUILD_ID" yaml:":buildkite_build_id"`
	BuildkiteBuildNumber            uint   `env:"BUILDKITE_BUILD_NUMBER" yaml:":buildkite_build_number"`
	BuildkiteBuildURL               string `env:"BUILDKITE_BUILD_URL" yaml:"-"`
	BuildkiteCommit                 string `env:"BUILDKITE_COMMIT" yaml:"-"`
	BuildkiteJobID                  string `env:"BUILDKITE_JOB_ID" yaml:":buildkite_job_id"`
	BuildkiteLabel                  string `env:"BUILDKITE_LABEL" yaml:":buildkite_label"`
	BuildkiteOrganizationSlug       string `env:"BUILDKITE_ORGANIZATION_SLUG" yaml:":buildkite_organization_slug"`
	BuildkitePipelineID             string `env:"BUILDKITE_PIPELINE_ID" yaml:":buildkite_pipeline_id"`
	BuildkitePipelineSlug           string `env:"BUILDKITE_PIPELINE_SLUG" yaml:":buildkite_pipeline_slug"`
	BuildkiteProjectSlug            string `env:"BUILDKITE_PROJECT_SLUG" yaml:":buildkite_project_slug"`
	BuildkitePullRequest            string `env:"BUILDKITE_PULL_REQUEST" yaml:"-"`
	BuildkitePullRequestBaseBranch  string `env:"BUILDKITE_PULL_REQUEST_BASE_BRANCH" yaml:":buildkite_pull_request_base_branch,omitempty"`
	BuildkitePullRequestNumber      uint   `yaml:":buildkite_pull_request_number,omitempty"`
	BuildkitePullRequestRepo        string `env:"BUILDKITE_PULL_REQUEST_REPO" yaml:":buildkite_pull_request_repo,omitempty"`
	BuildkiteRebuiltFromBuildID     string `env:"BUILDKITE_REBUILT_FROM_BUILD_ID" yaml:":buildkite_rebuilt_from_build_id,omitempty"`
	BuildkiteRebuiltFromBuildNumber uint   `env:"BUILDKITE_REBUILT_FROM_BUILD_NUMBER" yaml:":buildkite_rebuilt_from_build_number,omitempty"`
	BuildkiteRepoURL                string `env:"BUILDKITE_REPO" yaml:"-"`
	BuildkiteRetryCount             uint   `env:"BUILDKITE_RETRY_COUNT" yaml:":buildkite_retry_count"`
	BuildkiteTag                    string `env:"BUILDKITE_TAG" yaml:":buildkite_tag,omitempty"`
}

func (b *buildkiteMetadata) initEnvData(envs map[string]string, resolver CommitResolver) error {
	if err := env.Parse(b, env.Options{Environment: envs}); err != nil {
		return err
	}

	if err := b.initCommitData(resolver, b.BuildkiteCommit); err != nil {
		return err
	}

	b.Branch = b.BuildkiteBranch
	b.BuildURL = b.BuildkiteBuildURL
	b.CIProvider = "buildkite"

	nwo, err := nameWithOwnerFromGitURL(b.BuildkiteRepoURL)
	if err != nil {
		return err
	}
	b.RepoNameWithOwner = nwo

	prNum, err := strconv.ParseUint(b.BuildkitePullRequest, 0, 0)
	if err == nil {
		b.BuildkitePullRequestNumber = uint(prNum)
	}

	if b.Check == "" {
		b.Check = "buildkite"
	}

	return nil
}

func (b *buildkiteMetadata) MarshalYAML() (out []byte, err error) {
	return marshalYAML(b)
}

var _ Metadata = (*circleMetadata)(nil)

type circleMetadata struct {
	AbstractMetadata `yaml:",inline"`

	CircleBranch              string `env:"CIRCLE_BRANCH" yaml:"-"`
	CircleBuildNumber         uint   `env:"CIRCLE_BUILD_NUM" yaml:":circle_build_num"`
	CircleBuildURL            string `env:"CIRCLE_BUILD_URL" yaml:"-"`
	CircleJob                 string `env:"CIRCLE_JOB" yaml:":circle_job"`
	CircleProjectReponame     string `env:"CIRCLE_PROJECT_REPONAME" yaml:"-"`
	CircleProjectUsername     string `env:"CIRCLE_PROJECT_USERNAME" yaml:"-"`
	CirclePullRequestNumber   uint   `env:"CIRCLE_PR_NUMBER" yaml:":circle_pr_number,omitempty"`
	CirclePullRequestReponame string `env:"CIRCLE_PR_REPONAME" yaml:":circle_pr_reponame,omitempty"`
	CirclePullRequestURL      string `env:"CIRCLE_PULL_REQUEST" yaml:":circle_pull_request,omitempty"`
	CirclePullRequestUsername string `env:"CIRCLE_PR_USERNAME" yaml:":circle_pr_username,omitempty"`
	CircleRepoURL             string `env:"CIRCLE_REPOSITORY_URL" yaml:":circle_repository_url"`
	CircleSHA1                string `env:"CIRCLE_SHA1" yaml:"-"`
	CircleTag                 string `env:"CIRCLE_TAG" yaml:":circle_tag,omitempty"`
	CircleUsername            string `env:"CIRCLE_USERNAME" yaml:":circle_username"`
	CircleWorkflowID          string `env:"CIRCLE_WORKFLOW_ID" yaml:":circle_workflow_id"`
}

func (c *circleMetadata) initEnvData(envs map[string]string, resolver CommitResolver) error {
	if err := env.Parse(c, env.Options{Environment: envs}); err != nil {
		return err
	}

	if err := c.initCommitData(resolver, c.CircleSHA1); err != nil {
		return err
	}

	c.Branch = c.CircleBranch
	c.BuildURL = c.CircleBuildURL
	c.CIProvider = "circleci"
	c.RepoNameWithOwner = fmt.Sprintf("%s/%s", c.CircleProjectUsername, c.CircleProjectReponame)

	if c.Check == "" {
		c.Check = "circleci"
	}

	return nil
}

func (c *circleMetadata) MarshalYAML() (out []byte, err error) {
	return marshalYAML(c)
}

var _ Metadata = (*githubMetadata)(nil)

type githubMetadata struct {
	AbstractMetadata `yaml:",inline"`

	GithubActor     string `env:"GITHUB_ACTOR" yaml:":github_actor"`
	GithubBaseRef   string `env:"GITHUB_BASE_REF" yaml:":github_base_ref"`
	GithubHeadRef   string `env:"GITHUB_HEAD_REF" yaml:":github_head_ref"`
	GithubRef       string `env:"GITHUB_REF" yaml:":github_ref"`
	GithubRepoNWO   string `env:"GITHUB_REPOSITORY" yaml:"-"`
	GithubRepoURL   string `yaml:":github_repo_url"`
	GithubRunID     uint   `env:"GITHUB_RUN_ID" yaml:":github_run_id"`
	GithubRunNumber uint   `env:"GITHUB_RUN_NUMBER" yaml:":github_run_number"`
	GithubServerURL string `env:"GITHUB_SERVER_URL" yaml:"-"`
	GithubSHA       string `env:"GITHUB_SHA" yaml:"-"`
	GithubWorkflow  string `env:"GITHUB_WORKFLOW" yaml:":github_workflow"`
}

func (g *githubMetadata) initEnvData(envs map[string]string, resolver CommitResolver) error {
	if err := env.Parse(g, env.Options{Environment: envs}); err != nil {
		return err
	}

	if err := g.initCommitData(resolver, g.GithubSHA); err != nil {
		return err
	}

	g.RepoNameWithOwner = g.GithubRepoNWO
	g.GithubRepoURL = fmt.Sprintf("%s/%s", g.GithubServerURL, g.RepoNameWithOwner)
	g.BuildURL = fmt.Sprintf("%s/actions/runs/%d", g.GithubRepoURL, g.GithubRunID)
	g.CIProvider = "github-actions"

	branch, err := g.branch()
	if err != nil {
		return err
	}
	g.Branch = branch

	if g.Check == "" {
		g.Check = "github-actions"
	}

	return nil
}

func (g *githubMetadata) branch() (string, error) {
	isBranch, err := regexp.MatchString("^refs/heads/", g.GithubRef)
	if err != nil {
		return "", err
	}

	if !isBranch {
		return "", nil
	}

	return strings.TrimPrefix(g.GithubRef, "refs/heads/"), nil
}

func (g *githubMetadata) MarshalYAML() (out []byte, err error) {
	return marshalYAML(g)
}

var _ Metadata = (*jenkinsMetadata)(nil)

type jenkinsMetadata struct {
	AbstractMetadata `yaml:",inline"`

	GitBranch string `env:"GIT_BRANCH" yaml:"-"`
	GitCommit string `env:"GIT_COMMIT" yaml:"-"`
	GitURL    string `env:"GIT_URL" yaml:"-"`
}

func (j *jenkinsMetadata) initEnvData(envs map[string]string, resolver CommitResolver) error {
	if err := env.Parse(j, env.Options{Environment: envs}); err != nil {
		return err
	}

	if err := j.initCommitData(resolver, j.GitCommit); err != nil {
		return err
	}

	j.Branch = j.GitBranch
	j.CIProvider = "jenkins"

	url, ok := envs["BUILD_URL"]
	if !ok || url == "" {
		return fmt.Errorf("missing required environment variable: BUILD_URL")
	}
	j.BuildURL = url

	nwo, err := nameWithOwnerFromGitURL(j.GitURL)
	if err != nil {
		return err
	}
	j.RepoNameWithOwner = nwo

	if j.Check == "" {
		j.Check = "jenkins"
	}

	return nil
}

func (j *jenkinsMetadata) MarshalYAML() (out []byte, err error) {
	return marshalYAML(j)
}

var _ Metadata = (*semaphoreMetadata)(nil)

type semaphoreMetadata struct {
	AbstractMetadata `yaml:",inline"`

	SemaphoreAgentMachineEnvironmentType string `env:"SEMAPHORE_AGENT_MACHINE_ENVIRONMENT_TYPE" yaml:":semaphore_agent_machine_environment_type"`
	SemaphoreAgentMachineOsImage         string `env:"SEMAPHORE_AGENT_MACHINE_OS_IMAGE" yaml:":semaphore_agent_machine_os_image"`
	SemaphoreAgentMachineType            string `env:"SEMAPHORE_AGENT_MACHINE_TYPE" yaml:":semaphore_agent_machine_type"`
	SemaphoreGitBranch                   string `env:"SEMAPHORE_GIT_BRANCH" yaml:"-"`
	SemaphoreGitCommitRange              string `env:"SEMAPHORE_GIT_COMMIT_RANGE" yaml:":semaphore_git_commit_range"`
	SemaphoreGitDir                      string `env:"SEMAPHORE_GIT_DIR" yaml:":semaphore_git_dir"`
	SemaphoreGitRef                      string `env:"SEMAPHORE_GIT_REF" yaml:":semaphore_git_ref"`
	SemaphoreGitRefType                  string `env:"SEMAPHORE_GIT_REF_TYPE" yaml:":semaphore_git_ref_type"`
	SemaphoreGitRepoSlug                 string `env:"SEMAPHORE_GIT_REPO_SLUG" yaml:"-"`
	SemaphoreGitSHA                      string `env:"SEMAPHORE_GIT_SHA" yaml:"-"`
	SemaphoreGitURL                      string `env:"SEMAPHORE_GIT_URL" yaml:":semaphore_git_url"`
	SemaphoreJobID                       string `env:"SEMAPHORE_JOB_ID" yaml:":semaphore_job_id"`
	SemaphoreJobName                     string `env:"SEMAPHORE_JOB_NAME" yaml:":semaphore_job_name"`
	SemaphoreJobResult                   string `env:"SEMAPHORE_JOB_RESULT" yaml:":semaphore_job_result"`
	SemaphoreOrganizationURL             string `env:"SEMAPHORE_ORGANIZATION_URL" yaml:":semaphore_organization_url"`
	SemaphoreProjectID                   string `env:"SEMAPHORE_PROJECT_ID" yaml:":semaphore_project_id"`
	SemaphoreProjectName                 string `env:"SEMAPHORE_PROJECT_NAME" yaml:":semaphore_project_name"`
	SemaphoreWorkflowID                  string `env:"SEMAPHORE_WORKFLOW_ID" yaml:":semaphore_workflow_id"`
	SemaphoreWorkflowNumber              uint   `env:"SEMAPHORE_WORKFLOW_NUMBER" yaml:":semaphore_workflow_number"`
}

func (s *semaphoreMetadata) initEnvData(envs map[string]string, resolver CommitResolver) error {
	if err := env.Parse(s, env.Options{Environment: envs}); err != nil {
		return err
	}

	if err := s.initCommitData(resolver, s.SemaphoreGitSHA); err != nil {
		return err
	}

	s.Branch = s.SemaphoreGitBranch
	s.BuildURL = fmt.Sprintf("%s/workflows/%s", s.SemaphoreOrganizationURL, s.SemaphoreWorkflowID)
	s.CIProvider = "semaphore"
	s.RepoNameWithOwner = s.SemaphoreGitRepoSlug

	if s.Check == "" {
		s.Check = "semaphore"
	}

	return nil
}

func (s *semaphoreMetadata) MarshalYAML() (out []byte, err error) {
	return marshalYAML(s)
}

var _ Metadata = (*travisMetadata)(nil)

type travisMetadata struct {
	AbstractMetadata `yaml:",inline"`

	TravisBranch            string `env:"TRAVIS_BRANCH" yaml:"-"`
	TravisBuildDir          string `env:"TRAVIS_BUILD_DIR" yaml:":travis_build_dir"`
	TravisBuildID           uint   `env:"TRAVIS_BUILD_ID" yaml:":travis_build_id"`
	TravisBuildNumber       uint   `env:"TRAVIS_BUILD_NUMBER" yaml:":travis_build_number"`
	TravisBuildWebURL       string `env:"TRAVIS_BUILD_WEB_URL" yaml:":travis_build_web_url"`
	TravisCommit            string `env:"TRAVIS_COMMIT" yaml:"-"`
	TravisCommitRange       string `env:"TRAVIS_COMMIT_RANGE" yaml:":travis_commit_range"`
	TravisCPUArch           string `env:"TRAVIS_CPU_ARCH" yaml:":travis_cpu_arch"`
	TravisDist              string `env:"TRAVIS_DIST" yaml:":travis_dist"`
	TravisEventType         string `env:"TRAVIS_EVENT_TYPE" yaml:":travis_event_type"`
	TravisJobID             uint   `env:"TRAVIS_JOB_ID" yaml:":travis_job_id"`
	TravisJobName           string `env:"TRAVIS_JOB_NAME" yaml:":travis_job_name"`
	TravisJobNumber         string `env:"TRAVIS_JOB_NUMBER" yaml:":travis_job_number"`
	TravisJobWebURL         string `env:"TRAVIS_JOB_WEB_URL" yaml:"-"`
	TravisOsName            string `env:"TRAVIS_OS_NAME" yaml:":travis_os_name"`
	TravisPullRequest       string `env:"TRAVIS_PULL_REQUEST" yaml:"-"`
	TravisPullRequestBranch string `env:"TRAVIS_PULL_REQUEST_BRANCH" yaml:":travis_pull_request_branch,omitempty"`
	TravisPullRequestNumber uint   `yaml:":travis_pull_request_number,omitempty"`
	TravisPullRequestSha    string `env:"TRAVIS_PULL_REQUEST_SHA" yaml:":travis_pull_request_sha,omitempty"`
	TravisPullRequestSlug   string `env:"TRAVIS_PULL_REQUEST_SLUG" yaml:":travis_pull_request_slug,omitempty"`
	TravisRepoSlug          string `env:"TRAVIS_REPO_SLUG" yaml:"-"`
	TravisSudo              bool   `env:"TRAVIS_SUDO" yaml:":travis_sudo"`
	TravisTag               string `env:"TRAVIS_TAG" yaml:":travis_tag"`
	TravisTestResult        uint   `env:"TRAVIS_TEST_RESULT" yaml:":travis_test_result"`
}

func (t *travisMetadata) initEnvData(envs map[string]string, resolver CommitResolver) error {
	if err := env.Parse(t, env.Options{Environment: envs}); err != nil {
		return err
	}

	if err := t.initCommitData(resolver, t.TravisCommit); err != nil {
		return err
	}

	t.Branch = t.TravisBranch
	t.BuildURL = t.TravisJobWebURL
	t.CIProvider = "travis-ci"
	t.RepoNameWithOwner = t.TravisRepoSlug

	prNum, err := strconv.ParseUint(t.TravisPullRequest, 0, 0)
	if err == nil {
		t.TravisPullRequestNumber = uint(prNum)
	}

	if t.Check == "" {
		t.Check = "travis-ci"
	}

	return nil
}

func (t *travisMetadata) MarshalYAML() (out []byte, err error) {
	return marshalYAML(t)
}

func marshalYAML(m interface{}) (out []byte, err error) {
	data, err := yaml.Marshal(m)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func nameWithOwnerFromGitURL(url string) (string, error) {
	re := regexp.MustCompile(`github.com[:/](.*)`)

	matches := re.FindStringSubmatch(url)
	if len(matches) != 2 {
		return "", fmt.Errorf("unable to extract repository name-with-owner from URL: %s", url)
	}

	return strings.TrimSuffix(matches[1], ".git"), nil
}
