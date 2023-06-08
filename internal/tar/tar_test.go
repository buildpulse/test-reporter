package tar

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mholt/archiver/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTar(t *testing.T) {
	f, err := os.CreateTemp("", "*.tar")
	require.NoError(t, err)
	defer f.Close()

	tar := Create(f)

	err = tar.Write("./testdata/foo.txt", "foo.txt")
	require.NoError(t, err)

	err = tar.Write("./testdata/bar.TXT", "bar.TXT")
	require.NoError(t, err)

	err = tar.Write("./testdata/foo/bar/baz.txt", "foo/bar/baz.txt")
	require.NoError(t, err)

	err = tar.Write("./testdata/foo/bar/quux.txt", "foo/bar/quux.txt")
	require.NoError(t, err)

	err = tar.Close()
	require.NoError(t, err)

	untarDir := t.TempDir()
	err = archiver.Unarchive(f.Name(), untarDir)
	require.NoError(t, err)

	// === Verify resulting content is at expected location and matches original content
	assertEqualContent(t,
		"testdata/foo.txt",
		filepath.Join(untarDir, "foo.txt"),
	)
	assertEqualContent(t,
		"testdata/bar.TXT",
		filepath.Join(untarDir, "bar.TXT"),
	)
	assertEqualContent(t,
		"testdata/foo/bar/baz.txt",
		filepath.Join(untarDir, "foo/bar/baz.txt"),
	)
	assertEqualContent(t,
		"testdata/foo/bar/quux.txt",
		filepath.Join(untarDir, "foo/bar/quux.txt"),
	)
}

// assertEqualContent asserts that two files have the same content.
func assertEqualContent(t *testing.T, expected string, actual string) {
	expectedBytes, err := os.ReadFile(expected)
	require.NoError(t, err)

	actualBytes, err := os.ReadFile(actual)
	require.NoError(t, err)

	assert.Equal(t, expectedBytes, actualBytes)
}
