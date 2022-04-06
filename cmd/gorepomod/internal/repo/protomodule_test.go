// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package repo

import (
	"testing"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"sigs.k8s.io/kustomize/cmd/gorepomod/internal/misc"
)

func TestShortName(t *testing.T) {
	var testCases = map[string]struct {
		name    misc.ModuleShortName
		modFile *modfile.File
	}{
		"one": {
			name: misc.ModuleShortName("garage"),
			modFile: &modfile.File{
				Module: &modfile.Module{
					Mod: module.Version{
						Path:    "gh.com/micheal/garage",
						Version: "v2.3.4",
					},
				},
			},
		},
		"three": {
			name: misc.ModuleShortName("fruit/yellow/banana"),
			modFile: &modfile.File{
				Module: &modfile.Module{
					Mod: module.Version{
						Path:    "gh.com/micheal/fruit/yellow/banana",
						Version: "v2.3.4",
					},
				},
			},
		},
	}
	for n, tc := range testCases {
		m := protoModule{pathToGoMod: "irrelevant", mf: tc.modFile}
		actual := m.ShortName("gh.com/micheal")
		if actual != tc.name {
			t.Errorf(
				"%s: expected %s, got %s", n, tc.name, actual)
		}
	}
}
