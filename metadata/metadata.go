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
}

// AbstractMetadata provides the fields that are common across all Metadata
// instances, regardless of the specific CI provider.
type AbstractMetadata struct {
	Branch            string    `yaml:":branch"`
	BuildURL          string    `yaml:":build_url"`
	Check             string    `yaml:":check" env:"BUILDPULSE_CHECK_NAME"`
	CIProvider        string    `yaml:":ci_provider"`
	Commit            string    `yaml:":commit"`
	RepoNameWithOwner string    `yaml:":repo_name_with_owner"`
	Timestamp         time.Time `yaml:":timestamp"`
}

// NewMetadata creates a new Metadata instance from the given environment.
func NewMetadata(envs map[string]string, now func() time.Time) (Metadata, error) {
	switch {
	case envs["CIRCLECI"] == "true":
		return newCircleMetadata(envs, now)
	case envs["GITHUB_ACTIONS"] == "true":
		return newGithubMetadata(envs, now)
	case envs["SEMAPHORE"] == "true":
		return newSemaphoreMetadata(envs, now)
	case envs["TRAVIS"] == "true":
		return newTravisMetadata(envs, now)
	default:
		return nil, fmt.Errorf("unrecognized environment: system does not appear to be a supported CI provider (CircleCI, GitHub Actions, Semaphore, or Travis CI)")
	}
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

func newCircleMetadata(envs map[string]string, now func() time.Time) (Metadata, error) {
	m := &circleMetadata{}

	if err := env.Parse(m, env.Options{Environment: envs}); err != nil {
		return nil, err
	}

	m.Branch = m.CircleBranch
	m.BuildURL = m.CircleBuildURL
	m.CIProvider = "circleci"
	m.Commit = m.CircleSHA1
	m.RepoNameWithOwner = fmt.Sprintf("%s/%s", m.CircleProjectUsername, m.CircleProjectReponame)
	m.Timestamp = now()

	if m.Check == "" {
		m.Check = "circleci"
	}

	return m, nil
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

func newGithubMetadata(envs map[string]string, now func() time.Time) (Metadata, error) {
	m := &githubMetadata{}

	if err := env.Parse(m, env.Options{Environment: envs}); err != nil {
		return nil, err
	}

	m.RepoNameWithOwner = m.GithubRepoNWO
	m.GithubRepoURL = fmt.Sprintf("%s/%s", m.GithubServerURL, m.RepoNameWithOwner)
	m.BuildURL = fmt.Sprintf("%s/actions/runs/%d", m.GithubRepoURL, m.GithubRunID)
	m.CIProvider = "github-actions"
	m.Commit = m.GithubSHA
	m.Timestamp = now()

	branch, err := m.branch()
	if err != nil {
		return nil, err
	}
	m.Branch = branch

	if m.Check == "" {
		m.Check = "github-actions"
	}

	return m, nil
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

func newSemaphoreMetadata(envs map[string]string, now func() time.Time) (Metadata, error) {
	m := &semaphoreMetadata{}

	if err := env.Parse(m, env.Options{Environment: envs}); err != nil {
		return nil, err
	}

	m.Branch = m.SemaphoreGitBranch
	m.BuildURL = fmt.Sprintf("%s/workflows/%s", m.SemaphoreOrganizationURL, m.SemaphoreWorkflowID)
	m.CIProvider = "semaphore"
	m.Commit = m.SemaphoreGitSHA
	m.RepoNameWithOwner = m.SemaphoreGitRepoSlug
	m.Timestamp = now()

	if m.Check == "" {
		m.Check = "semaphore"
	}

	return m, nil
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

func newTravisMetadata(envs map[string]string, now func() time.Time) (Metadata, error) {
	m := &travisMetadata{}

	if err := env.Parse(m, env.Options{Environment: envs}); err != nil {
		return nil, err
	}

	m.Branch = m.TravisBranch
	m.BuildURL = m.TravisJobWebURL
	m.CIProvider = "travis-ci"
	m.Commit = m.TravisCommit
	m.RepoNameWithOwner = m.TravisRepoSlug
	m.Timestamp = now()

	prNum, err := strconv.ParseUint(m.TravisPullRequest, 0, 0)
	if err == nil {
		m.TravisPullRequestNumber = uint(prNum)
	}

	if m.Check == "" {
		m.Check = "travis-ci"
	}

	return m, nil
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
