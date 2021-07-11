package tar

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
)

// A Tar instance provides functionality for creating a tar archive.
type Tar struct {
	dirs   map[string]struct{} // the keys of this map represent the set of directory paths that have been written to the tar
	writer *tar.Writer
}

// Create opens t for writing a tar archive to out.
func Create(out io.Writer) *Tar {
	t := &Tar{
		dirs:   make(map[string]struct{}),
		writer: tar.NewWriter(out),
	}

	return t
}

// Write writes the file at src into t at the given dest path.
func (t *Tar) Write(src string, dest string) error {
	// Write a header for the directory containing this file (if we haven't already done so)
	destdir := filepath.Dir(dest)
	_, ok := t.dirs[destdir]
	if !ok && destdir != "." {
		dirinfo, err := os.Lstat(filepath.Dir(src))
		if err != nil {
			return err
		}

		err = t.writeHeader(dirinfo, destdir)
		if err != nil {
			return err
		}

		t.dirs[destdir] = struct{}{}
	}

	info, err := os.Lstat(src)
	if err != nil {
		return err
	}

	err = t.writeHeader(info, dest)
	if err != nil {
		return err
	}

	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(t.writer, file)
	return err
}

// Close closes the archive opened by Create.
func (t *Tar) Close() error {
	if t.writer != nil {
		w := t.writer
		t.writer = nil

		err := w.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Tar) writeHeader(info os.FileInfo, dest string) error {
	header, err := tar.FileInfoHeader(info, dest)
	if err != nil {
		return err
	}

	header.Name = dest
	if err := t.writer.WriteHeader(header); err != nil {
		return err
	}

	return nil
}
