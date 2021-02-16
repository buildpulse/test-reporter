package metadata

import (
	"path"
	"testing"
	"time"

	"github.com/otiai10/copy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRepositoryCommitResolver_invalidRepo(t *testing.T) {
	_, err := NewRepositoryCommitResolver(t.TempDir())
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "no repository found")
	}
}

func Test_repositoryCommitResolver_Lookup(t *testing.T) {
	dir := t.TempDir()
	err := copy.Copy("./testdata/example-repository.git", path.Join(dir, ".git"))
	require.NoError(t, err)

	r, err := NewRepositoryCommitResolver(dir)
	require.NoError(t, err)

	c, err := r.Lookup("5974e4edce87279f60adaf55c2adcee8847b2612")
	require.NoError(t, err)

	assert.Equal(t, "Thu Dec 31 01:02:03 +1300 2020", c.AuthoredAt.Format(time.UnixDate))
	assert.Equal(t, "dwight@dundermifflin.com", c.AuthorEmail)
	assert.Equal(t, "Dwight Kurt Schrute III", c.AuthorName)
	assert.Equal(t, "Fri Jan  1 13:14:15 -0800 2021", c.CommittedAt.Format(time.UnixDate))
	assert.Equal(t, "jim@dundermifflin.com", c.CommitterEmail)
	assert.Equal(t, "Jim Halpert", c.CommitterName)
	assert.Equal(t, "Add some flair\n", c.Message)
	assert.Equal(t, "5974e4edce87279f60adaf55c2adcee8847b2612", c.SHA)
	assert.Equal(t, "eb8b39c87131c1f3543bc6e5a426f7d4d631bc15", c.TreeSHA)
}

func Test_repositoryCommitResolver_Lookup_notFound(t *testing.T) {
	dir := t.TempDir()
	err := copy.Copy("./testdata/example-repository.git", path.Join(dir, ".git"))
	require.NoError(t, err)

	r, err := NewRepositoryCommitResolver(dir)
	require.NoError(t, err)

	_, err = r.Lookup("0000000000000000000000000000000000000000")
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "unable to find commit with SHA 0000000000000000000000000000000000000000")
	}
}
