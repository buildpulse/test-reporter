package metadata

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// Commit represents the metadata for a Git commit.
type Commit struct {
	SHA     string
	TreeSHA string
}

// A CommitResolver provides the ability to look up a commit.
type CommitResolver interface {
	Lookup(sha string) (*Commit, error)
}

// The CommitResolverFunc type is an adapter to allow the use of ordinary
// functions as commit resolvers. If f is a function with the appropriate
// signature, CommitResolverFunc(f) is a CommitResolver that calls f.
type CommitResolverFunc func(sha string) (*Commit, error)

// Lookup calls f(sha).
func (f CommitResolverFunc) Lookup(sha string) (*Commit, error) {
	return f(sha)
}

// NewCommitResolver returns a CommitResolver for looking up commits in the
// repository located at path.
//
// Future CommitResolver variants could potentially include the ability to
// lookup a commit based on environment variables or args passed to the
// test-reporter CLI.
func NewCommitResolver(path string) (CommitResolver, error) {
	repo, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		if err == git.ErrRepositoryNotExists {
			return nil, fmt.Errorf("no repository found at %s", path)
		}

		return nil, err
	}

	return &repositoryCommitResolver{repo: repo}, nil
}

type repositoryCommitResolver struct {
	repo *git.Repository
}

func (r *repositoryCommitResolver) Lookup(sha string) (*Commit, error) {
	c, err := r.repo.CommitObject(plumbing.NewHash(sha))
	if err != nil {
		return nil, fmt.Errorf("unable to find commit with SHA %s: %v", sha, err)
	}

	return &Commit{
		SHA:     sha,
		TreeSHA: c.TreeHash.String(),
	}, nil
}
