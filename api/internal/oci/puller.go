// Copyright 2025 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// Puller is a function that can clone a git repo.
type Puller func(repoSpec *RepoSpec, fSys filesys.FileSystem, client *http.Client) error

// PullUsingOciManifest pulls/copies the oci manifest definition
func PullUsingOciManifest(repoSpec *RepoSpec, fSys filesys.FileSystem, client *http.Client) error {
	ctx := context.Background()

	transport := http.DefaultTransport
	if client != nil && client.Transport != nil {
		transport = client.Transport
	}

	image, err := remote.Image(repoSpec.Reference, remote.WithTransport(transport), remote.WithContext(ctx))
	if err != nil {
		return err
	}

	// Using extraction to get the "flattened" image until https://github.com/google/go-containerregistry/issues/921 is implemented
	extracted := mutate.Extract(image)
	defer extracted.Close()

	reader := tar.NewReader(extracted)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}

		// Don't care about empty directories
		if header.FileInfo().IsDir() {
			continue
		}

		if !filepath.IsLocal(header.Name) {
			return fmt.Errorf("file reference '%s' not relative", header.Name)
		}

		destination := repoSpec.Dir.Join(header.Name)
		directory := filepath.Dir(destination)

		if err := fSys.MkdirAll(directory); err != nil {
			return err
		}

		if fp, err := fSys.Open(destination); err != nil {
			fp.Close()
			return err
		} else if _, err := io.Copy(fp, reader); err != nil {
			fp.Close()
			return err
		} else {
			fp.Close()
		}
	}

	return nil
}

// DoNothingPuller returns a puller that only sets
// pullDir field in the repoSpec.  It's assumed that
// the pullDir is associated with some fake filesystem
// used in a test.
func DoNothingPuller(dir filesys.ConfirmedDir) Puller {
	return func(rs *RepoSpec, fSys filesys.FileSystem, client *http.Client) error {
		rs.Dir = dir
		return nil
	}
}
