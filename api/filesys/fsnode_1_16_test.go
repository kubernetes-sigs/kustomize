// +build 1.16

package filesys

import (
	"fmt"
	"strings"
	"testing"
	"testing/iotest"
)

func TestFileOps(t *testing.T) {
	const path = "foo.txt"
	content := strings.Repeat("longest content", 100)

	fs := MakeFsInMemory()
	f, err := fs.Create(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := fs.Open(path); err == nil {
		t.Fatalf("expected already opened error, got nil")
	}
	if _, err := fmt.Fprint(f, content); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := f.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := f.Close(); err == nil {
		t.Fatalf("expected already closed error, got nil")
	}

	f, err = fs.Open(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()

	if err := iotest.TestReader(f, []byte(content)); err != nil {
		t.Fatalf("test reader failed: %v", err)
	}
}
