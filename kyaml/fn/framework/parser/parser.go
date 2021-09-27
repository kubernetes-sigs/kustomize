// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	iofs "io/fs"
	"io/ioutil"
	"path"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/errors"
)

type parser struct {
	fs        iofs.FS
	paths     []string
	extension string
}

type contentProcessor func(content []byte, name string) error

func (l parser) parse(processContent contentProcessor) error {
	for _, path := range l.paths {
		if err := l.readPath(path, processContent); err != nil {
			return err
		}
	}
	return nil
}

func (l parser) readPath(path string, processContent contentProcessor) error {
	f, err := l.fs.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	// File is directory -- read templates among its immediate children
	if info.IsDir() {
		dir, ok := f.(iofs.ReadDirFile)
		if !ok {
			return errors.Errorf("%s is a directory but could not be opened as one", path)
		}
		return l.readDir(dir, path, processContent)
	}

	// Path is a file -- check extension and read it
	if !strings.HasSuffix(path, l.extension) {
		return errors.Errorf("file %s did not have required extension %s", path, l.extension)
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	return processContent(b, path)
}

func (l parser) readDir(dir iofs.ReadDirFile, dirname string, processContent contentProcessor) error {
	entries, err := dir.ReadDir(0)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), l.extension) {
			continue
		}
		// Note: using filepath.Join will break Windows, because io/fs.FS implementations require slashes on all OS.
		// See https://golang.org/pkg/io/fs/#ValidPath
		b, err := l.readFile(path.Join(dirname, entry.Name()))
		if err != nil {
			return err
		}
		if err := processContent(b, entry.Name()); err != nil {
			return err
		}
	}
	return nil
}

func (l parser) readFile(path string) ([]byte, error) {
	f, err := l.fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ioutil.ReadAll(f)
}
