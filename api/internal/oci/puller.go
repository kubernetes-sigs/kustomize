// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package oci

import (
	"archive/tar"
	"io"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/authn"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// Puller is a function that can pull an OCI artifact and
// extract its contents into repoSpec.Dir on the given file system.
type Puller func(repoSpec *RepoSpec, fSys filesys.FileSystem) error

// PullerUsingRegistry pulls the artifact from its registry, using the
// credentials of the local container tooling (e.g. ~/.docker/config.json)
// when the registry requires authentication, and extracts the flattened
// layers into a new temporary directory.
func PullerUsingRegistry(repoSpec *RepoSpec, fSys filesys.FileSystem) error {
	dir, err := filesys.NewTmpConfirmedDir()
	if err != nil {
		return errors.Wrap(err)
	}
	// Set before anything can fail, so that the cleaner removes the directory.
	repoSpec.Dir = dir

	img, err := remote.Image(repoSpec.Reference,
		remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return errors.WrapPrefixf(err, "pulling oci artifact %s", repoSpec.PullSpec())
	}
	return extract(img, dir, fSys)
}

// extract writes the flattened file system of the image below dir.
// Layer media types are not inspected: each layer is expected to be a tar
// archive, optionally compressed, which covers container images as well as
// artifacts pushed by tools like flux or oras.
func extract(img v1.Image, dir filesys.ConfirmedDir, fSys filesys.FileSystem) error {
	flattened := mutate.Extract(img)
	defer flattened.Close()

	reader := tar.NewReader(flattened)
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			// The tar trailer is written before an error from Extract
			// closes the stream; drain it so that the error surfaces.
			if _, err = io.Copy(io.Discard, flattened); err != nil {
				return errors.WrapPrefixf(err, "reading contents of oci artifact")
			}
			return nil
		}
		if err != nil {
			return errors.WrapPrefixf(err, "reading contents of oci artifact")
		}
		if !filepath.IsLocal(header.Name) {
			return errors.Errorf(
				"oci artifact entry %q is not local to the artifact", header.Name)
		}
		target := dir.Join(header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err = fSys.MkdirAll(target); err != nil {
				return errors.Wrap(err)
			}
		case tar.TypeReg:
			if err = writeFile(fSys, target, reader); err != nil {
				return err
			}
		default:
			// Links, devices and the like are not meaningful in a
			// kustomization artifact, and links could be used to
			// reach outside of the pull directory; skip them.
		}
	}
}

func writeFile(fSys filesys.FileSystem, target string, content io.Reader) error {
	if err := fSys.MkdirAll(filepath.Dir(target)); err != nil {
		return errors.Wrap(err)
	}
	data, err := io.ReadAll(content)
	if err != nil {
		return errors.WrapPrefixf(err, "reading %s from oci artifact", target)
	}
	return errors.Wrap(fSys.WriteFile(target, data))
}

// DoNothingPuller returns a puller that only sets the
// Dir field in the repoSpec. It's assumed that the
// dir is associated with some fake filesystem
// used in a test.
func DoNothingPuller(dir filesys.ConfirmedDir) Puller {
	return func(rs *RepoSpec, _ filesys.FileSystem) error {
		rs.Dir = dir
		return nil
	}
}
