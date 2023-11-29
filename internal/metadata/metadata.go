package metadata

import (
	"strings"
	"time"

	"github.com/buildpulse/test-reporter/internal/logger"
	"gopkg.in/yaml.v3"
)

// A Metadata instance provides metadata about a set of test results. It
// identifies the CI provider, the commit SHA, the time at which the tests were
// executed, etc.
type Metadata struct {
	AuthoredAt           time.Time `yaml:":authored_at,omitempty"`
	AuthorEmail          string    `yaml:":author_email,omitempty"`
	AuthorName           string    `yaml:":author_name,omitempty"`
	Branch               string    `yaml:":branch"`
	BuildURL             string    `yaml:":build_url"`
	Check                string    `yaml:":check"`
	CIProvider           string    `yaml:":ci_provider"`
	CommitMessage        string    `yaml:":commit_message,omitempty"`
	CommitMetadataSource string    `yaml:":commit_metadata_source"`
	CommitSHA            string    `yaml:":commit"`
	CommittedAt          time.Time `yaml:":committed_at,omitempty"`
	CommitterEmail       string    `yaml:":committer_email,omitempty"`
	CommitterName        string    `yaml:":committer_name,omitempty"`
	QuotaID              string    `yaml:":quota_id,omitempty"`
	RepoNameWithOwner    string    `yaml:":repo_name_with_owner"`
	ReporterOS           string    `yaml:":reporter_os"`
	ReporterVersion      string    `yaml:":reporter_version"`
	Tags                 []string  `yaml:":tags,omitempty"`
	Timestamp            time.Time `yaml:":timestamp"`
	TreeSHA              string    `yaml:":tree,omitempty"`

	logger       logger.Logger
	providerData providerMetadata
}

// NewMetadata creates a new Metadata instance from the given args.
func NewMetadata(version *Version, envs map[string]string, tags []string, quotaID string, resolver CommitResolver, now func() time.Time, logger logger.Logger) (*Metadata, error) {
	m := &Metadata{logger: logger}

	if err := m.initProviderData(envs); err != nil {
		return nil, err
	}

	if err := m.initCommitData(resolver, m.providerData.CommitSHA()); err != nil {
		return nil, err
	}

	m.initTimestamp(now)
	m.initVersionData(version)

	m.QuotaID = quotaID
	m.Tags = tags

	return m, nil
}

func (m *Metadata) initProviderData(envs map[string]string) error {
	pm, err := newProviderMetadata(envs, m.logger)
	if err != nil {
		return err
	}
	m.providerData = pm

	m.Branch = pm.Branch()
	m.BuildURL = pm.BuildURL()
	m.CIProvider = pm.Name()
	m.RepoNameWithOwner = pm.RepoNameWithOwner()

	check, ok := envs["BUILDPULSE_CHECK_NAME"]
	if ok && check != "" {
		m.Check = check
	} else {
		m.Check = pm.Name()
	}

	return nil
}

func (m *Metadata) initCommitData(cr CommitResolver, sha string) error {
	m.CommitMetadataSource = cr.Source()

	c, err := cr.Lookup(sha)
	if err != nil {
		m.logger.Printf("❌")
		m.logger.Printf("❌ Commit lookup unsuccessful: %v", err)
		m.logger.Printf("❌")
		m.logger.Printf("❌ Test results will not be analyzed for this build. Please get in touch at https://buildpulse.io/contact so we can resolve this problem together.")
		m.logger.Printf("❌")
		m.logger.Printf("❌ In a future release, this issue will become a fatal error with a nonzero exit code.")
		m.logger.Printf("❌")

		m.CommitSHA = sha

		return nil
	}

	m.AuthoredAt = c.AuthoredAt
	m.AuthorEmail = c.AuthorEmail
	m.AuthorName = c.AuthorName
	m.CommitMessage = strings.TrimSpace(c.Message)
	m.CommitSHA = c.SHA
	m.CommittedAt = c.CommittedAt
	m.CommitterEmail = c.CommitterEmail
	m.CommitterName = c.CommitterName
	m.TreeSHA = c.TreeSHA

	return nil
}

func (m *Metadata) initTimestamp(now func() time.Time) {
	m.Timestamp = now()
}

func (m *Metadata) initVersionData(version *Version) {
	m.ReporterOS = version.GoOS
	m.ReporterVersion = version.Number
}

// MarshalYAML serializes the metadata into a YAML document.
func (m *Metadata) MarshalYAML() (out []byte, err error) {
	universalFields, err := yaml.Marshal(m)
	if err != nil {
		return nil, err
	}

	providerSpecificFields, err := yaml.Marshal(m.providerData)
	if err != nil {
		return nil, err
	}

	return append(universalFields, providerSpecificFields...), nil
}
