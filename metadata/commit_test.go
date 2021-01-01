package metadata

import (
	"path"
	"testing"
	"time"

	"github.com/otiai10/copy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCommitResolver_invalidRepo(t *testing.T) {
	_, err := NewCommitResolver(t.TempDir())
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "no repository found")
	}
}

func Test_repositoryCommitResolver_Lookup(t *testing.T) {
	dir := t.TempDir()
	err := copy.Copy("./testdata/example-repository.git", path.Join(dir, ".git"))
	require.NoError(t, err)

	r, err := NewCommitResolver(dir)
	require.NoError(t, err)

	c, err := r.Lookup("616a8b20906bd3820fb489148ed69516be3b98ad")
	require.NoError(t, err)

	assert.Equal(t, "Thu Dec 31 01:02:03 -0500 2020", c.AuthoredAt.Format(time.UnixDate))
	assert.Equal(t, "dwight@dundermifflin.com", c.AuthorEmail)
	assert.Equal(t, "Dwight Kurt Schrute III", c.AuthorName)
	assert.Equal(t, "Fri Jan  1 13:14:15 -0500 2021", c.CommittedAt.Format(time.UnixDate))
	assert.Equal(t, "jim@dundermifflin.com", c.CommitterEmail)
	assert.Equal(t, "Jim Halpert", c.CommitterName)
	assert.Equal(t, "Add some flair\n", c.Message)
	assert.Equal(t, "616a8b20906bd3820fb489148ed69516be3b98ad", c.SHA)
	assert.Equal(t, "eb8b39c87131c1f3543bc6e5a426f7d4d631bc15", c.TreeSHA)

}

func Test_repositoryCommitResolver_Lookup_notFound(t *testing.T) {
	dir := t.TempDir()
	err := copy.Copy("./testdata/example-repository.git", path.Join(dir, ".git"))
	require.NoError(t, err)

	r, err := NewCommitResolver(dir)
	require.NoError(t, err)

	_, err = r.Lookup("0000000000000000000000000000000000000000")
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "unable to find commit with SHA 0000000000000000000000000000000000000000")
	}
}
