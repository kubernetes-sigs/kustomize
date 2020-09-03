// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package loader

import (
	"context"
	"log"
	"os"

	"github.com/yujunz/go-getter"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/internal/git"
)

type remoteTargetSpec struct {
	// Raw is the original resource in kustomization.yaml
	Raw string

	// Dir is where the resource is saved
	Dir filesys.ConfirmedDir

	// TempDir is the directory created to hold all resources, including Dir
	TempDir filesys.ConfirmedDir
}

// Getter is a function that can gets resource
type remoteTargetGetter func(rs *remoteTargetSpec) error

func newLoaderAtGetter(raw string, fSys filesys.FileSystem, referrer *fileLoader, cloner git.Cloner, getter remoteTargetGetter) (ifc.Loader, error) {
	rs := &remoteTargetSpec{
		Raw: raw,
	}

	cleaner := func() error {
		return fSys.RemoveAll(rs.TempDir.String())
	}

	if err := getter(rs); err != nil {
		cleaner()
		return nil, err
	}

	return &fileLoader{
		loadRestrictor: RestrictionRootOnly,
		// TODO(yujunz): limit to getter root
		root:     rs.Dir,
		referrer: referrer,
		fSys:     fSys,
		cloner:   cloner,
		rscSpec:  rs,
		getter:   getter,
		cleaner:  cleaner,
	}, nil
}

func getRemoteTarget(rs *remoteTargetSpec) error {
	var err error

	rs.TempDir, err = filesys.NewTmpConfirmedDir()
	if err != nil {
		return err
	}

	rs.Dir = filesys.ConfirmedDir(rs.TempDir.Join("repo"))

	// Get the pwd
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting wd: %s", err)
	}

	opts := []getter.ClientOption{}
	client := &getter.Client{
		Ctx:  context.TODO(),
		Src:  rs.Raw,
		Dst:  rs.Dir.String(),
		Pwd:  pwd,
		Mode: getter.ClientModeAny,
		Detectors: []getter.Detector{
			new(getter.GitHubDetector),
			new(getter.GitDetector),
			new(getter.BitBucketDetector),
		},
		Options: opts,
	}
	return client.Get()
}

func getNothing(rs *remoteTargetSpec) error {
	var err error
	rs.Dir, err = filesys.NewTmpConfirmedDir()
	if err != nil {
		return err
	}

	// Get the pwd
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting wd: %s", err)
	}

	_, err = getter.Detect(rs.Raw, pwd, []getter.Detector{})
	return err
}
