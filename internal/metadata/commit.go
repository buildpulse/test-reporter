package metadata

import (
	"fmt"
	"time"

	"github.com/buildpulse/test-reporter/internal/logger"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// Commit represents the metadata for a Git commit.
type Commit struct {
	AuthoredAt     time.Time
	AuthorEmail    string
	AuthorName     string
	CommittedAt    time.Time
	CommitterEmail string
	CommitterName  string
	Message        string
	SHA            string
	TreeSHA        string
}

// A CommitResolver provides the ability to look up a commit.
type CommitResolver interface {
	Lookup(sha string) (*Commit, error)
	Source() string
}

type repositoryCommitResolver struct {
	logger logger.Logger
	repo   *git.Repository
}

// NewRepositoryCommitResolver returns a CommitResolver for looking up commits
// in the repository located at path.
func NewRepositoryCommitResolver(path string, logger logger.Logger) (CommitResolver, error) {
	repo, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		if err == git.ErrRepositoryNotExists {
			return nil, fmt.Errorf("no repository found at %s", path)
		}

		return nil, err
	}

	return &repositoryCommitResolver{repo: repo, logger: logger}, nil
}

func (r *repositoryCommitResolver) Lookup(sha string) (*Commit, error) {
	r.logger.Printf("Looking up info for commit `%s` in git repository", sha)
	c, err := r.repo.CommitObject(plumbing.NewHash(sha))
	if err != nil {
		// To help with diagnosing this error, try to log the HEAD reference, but if we encounter an error, just move on.
		head, headErr := r.repo.Head()
		if headErr == nil {
			r.logger.Printf("Repository's HEAD reference is %s", head)
		}

		return nil, fmt.Errorf("unable to find commit with SHA `%s`: %v", sha, err)
	}
	r.logger.Println("Found commit info")

	return &Commit{
		AuthoredAt:     c.Author.When,
		AuthorEmail:    c.Author.Email,
		AuthorName:     c.Author.Name,
		CommittedAt:    c.Committer.When,
		CommitterEmail: c.Committer.Email,
		CommitterName:  c.Committer.Name,
		Message:        c.Message,
		SHA:            c.Hash.String(),
		TreeSHA:        c.TreeHash.String(),
	}, nil
}

func (r *repositoryCommitResolver) Source() string {
	return "Repository"
}

type staticCommitResolver struct {
	commit *Commit
}

// NewStaticCommitResolver returns a CommitResolver whose Lookup method always
// produces a Commit with values matching the fields in c.
func NewStaticCommitResolver(c *Commit, logger logger.Logger) CommitResolver {
	return &staticCommitResolver{commit: c}
}

func (s *staticCommitResolver) Lookup(sha string) (*Commit, error) {
	return &Commit{
		SHA:            sha,
		AuthoredAt:     s.commit.AuthoredAt,
		AuthorEmail:    s.commit.AuthorEmail,
		AuthorName:     s.commit.AuthorName,
		CommittedAt:    s.commit.CommittedAt,
		CommitterEmail: s.commit.CommitterEmail,
		CommitterName:  s.commit.CommitterName,
		Message:        s.commit.Message,
		TreeSHA:        s.commit.TreeSHA,
	}, nil
}

func (s *staticCommitResolver) Source() string {
	return "Static"
}
