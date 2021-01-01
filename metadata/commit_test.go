package metadata

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/otiai10/copy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCommitResolver_invalidRepo(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "buildpulse-new-commit-resolver-")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	_, err = NewCommitResolver(dir)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "no repository found")
	}
}

func Test_repositoryCommitResolver_Lookup(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "buildpulse-repository-commit-resolver-")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	err = copy.Copy("./testdata/example-repository.git", path.Join(dir, ".git"))
	require.NoError(t, err)

	r, err := NewCommitResolver(dir)
	require.NoError(t, err)

	c, err := r.Lookup("e1fe5ac0f2dced788af7669f242dd74317a8d0e0")
	require.NoError(t, err)
	assert.Equal(t, "e1fe5ac0f2dced788af7669f242dd74317a8d0e0", c.SHA)
	assert.Equal(t, "7a5800f1746bc4ab0f7dd3297d98d363cccc4347", c.TreeSHA)
}

func Test_repositoryCommitResolver_Lookup_notFound(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "buildpulse-repository-commit-resolver-")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	err = copy.Copy("./testdata/example-repository.git", path.Join(dir, ".git"))
	require.NoError(t, err)

	r, err := NewCommitResolver(dir)
	require.NoError(t, err)

	_, err = r.Lookup("0000000000000000000000000000000000000000")
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "unable to find commit with SHA 0000000000000000000000000000000000000000")
	}
}
