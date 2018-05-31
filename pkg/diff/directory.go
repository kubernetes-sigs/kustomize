package diff

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// directory represents a new temp directory and lets one create files in it.
type directory struct {
	n string
}

// newDirectory makes a directory instance holding a new temp directory on disk.
// The directory name is the given prefix following by a random string.
func newDirectory(prefix string) (*directory, error) {
	name, err := ioutil.TempDir("", prefix+"-")
	if err != nil {
		return nil, err
	}

	return &directory{n: name}, nil
}

// newFile creates a new file in the directory.
func (d *directory) newFile(name string) (*os.File, error) {
	return os.OpenFile(filepath.Join(d.n, name), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0700)
}

// delete removes the directory recursively.
func (d *directory) delete() error {
	return os.RemoveAll(d.n)
}

// name is the name of the directory.
func (d *directory) name() string {
	return d.n
}
