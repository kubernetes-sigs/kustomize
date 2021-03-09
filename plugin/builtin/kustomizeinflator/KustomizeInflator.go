// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate pluginator
package main

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	GlobPrefix = "glob:"
)

type plugin struct {
	KustomizationPaths []string `yaml:"kustomizationPaths"`
	// h contains plugin helpers
	h *resmap.PluginHelpers
}

//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

func (ki *plugin) Config(h *resmap.PluginHelpers, config []byte) error {

	err := yaml.Unmarshal(config, ki)
	if err != nil {
		return err
	}
	ki.h = h
	return nil
}

func (ki *plugin) Generate() (resmap.ResMap, error) {
	paths, err := expandKustomizationPaths(ki.KustomizationPaths, ki.h.Loader().Root())
	if err != nil {
		return nil, fmt.Errorf("Error expanding kustomization paths: %v", err)
	}
	resMap := resmap.New()
	for _, p := range paths {
		log.Debugf("Inflating kustomization `%s'", p)
		args := []string{"build", "--load_restrictor", "none",
			"--enable_alpha_plugins", p}
		cmd := exec.Command("kustomize", args...)
		out, err := cmd.Output()
		if err != nil {
			msg := fmt.Sprintf("Error inflating kustomization `%s': %v.", p, err)
			var exitError *exec.ExitError
			if errors.As(err, &exitError) {
				msg = fmt.Sprintf("%s Stderr: '%s'", msg, string(exitError.Stderr))
			}
			return nil, errors.New(msg)
		}

		kustResMap, err := ki.h.ResmapFactory().NewResMapFromBytes(out)
		if err != nil {
			return nil, fmt.Errorf("Failed to create ResMap for kustomization: %v", err)
		}
		err = resMap.AbsorbAll(kustResMap)
		if err != nil {
			return nil, err
		}
	}
	return resMap, nil
}

func expandKustomizationPaths(paths []string, kustPath string) ([]string, error) {
	res := []string{}
	for _, p := range paths {
		// Handle glob. If the glob returns no matches, continue as normal.
		// This behavior is similar to bash's nullglob option:
		// https://www.gnu.org/software/bash/manual/html_node/The-Shopt-Builtin.html
		if strings.HasPrefix(p, GlobPrefix) {
			glob := strings.TrimPrefix(p, GlobPrefix)
			glob = filepath.Join(kustPath, glob)
			pathExpasion, err := filepath.Glob(glob)
			if err != nil {
				return nil, fmt.Errorf("Error expanding glob: %w", err)
			}
			res = append(res, pathExpasion...)
		} else {
			// Handle normal path
			p = filepath.Join(kustPath, p)
			res = append(res, p)
		}
	}
	return res, nil
}
