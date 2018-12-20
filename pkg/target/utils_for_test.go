/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package target

// A collection of utilities used in target tests.

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/pkg/constants"
	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/internal/loadertest"
)

func makeKustTarget(t *testing.T, l ifc.Loader) *KustTarget {
	// Warning: the following filesystem - a fake - must be rooted at /.
	// This fs root is used as the working directory for the shell spawned by
	// the secretgenerator, and has nothing to do with the filesystem used
	// to load relative paths from the fake filesystem.
	// This trick only works for secret generator commands that don't actually
	// try to read the file system, because these tests don't write to the
	// real "/" directory.  See use of exec package in the secretfactory.
	fakeFs := fs.MakeFakeFS()
	fakeFs.Mkdir("/")
	kt, err := NewKustTarget(
		l, fakeFs, rf, transformer.NewFactoryImpl())
	if err != nil {
		t.Fatalf("Unexpected construction error %v", err)
	}
	return kt
}

func writeF(
	t *testing.T, ldr loadertest.FakeLoader, dir string, content string) {
	err := ldr.AddFile(dir, []byte(content))
	if err != nil {
		t.Fatalf("failed write to %s; %v", dir, err)
	}
}

func writeK(
	t *testing.T, ldr loadertest.FakeLoader, dir string, content string) {
	writeF(t, ldr, filepath.Join(dir, constants.KustomizationFileName), `
apiVersion: v1beta1
kind: Kustomization
`+content)
}

func tabToSpace(input string) string {
	var result []string
	for _, i := range input {
		if i == 9 {
			result = append(result, "  ")
		} else {
			result = append(result, string(i))
		}
	}
	return strings.Join(result, "")
}

func convertToArray(x string) ([]string, int) {
	a := strings.Split(strings.TrimSuffix(x, "\n"), "\n")
	maxLen := 0
	for i, v := range a {
		z := tabToSpace(v)
		if len(z) > maxLen {
			maxLen = len(z)
		}
		a[i] = z
	}
	return a, maxLen
}

func hint(a, b string) string {
	if a == b {
		return " "
	}
	return "X"
}

// Pretty printing of file differences.
func assertExpectedEqualsActual(t *testing.T, actual []byte, expected string) {
	if expected == string(actual) {
		return
	}
	sE, maxLen := convertToArray(expected)
	sA, _ := convertToArray(string(actual))
	format := fmt.Sprintf("%%s  %%-%ds %%s\n", maxLen+4)

	limit := 0
	if len(sE) < len(sA) {
		limit = len(sE)
	} else {
		limit = len(sA)
	}

	fmt.Printf(format, " ", "EXPECTED", "ACTUAL")
	fmt.Printf(format, " ", "--------", "------")
	for i := 0; i < limit; i++ {
		fmt.Printf(format, hint(sE[i], sA[i]), sE[i], sA[i])
	}
	if len(sE) < len(sA) {
		for i := len(sE); i < len(sA); i++ {
			fmt.Printf(format, "X", "", sA[i])
		}
	} else {
		for i := len(sA); i < len(sE); i++ {
			fmt.Printf(format, "X", sE[i], "")
		}
	}
	t.Fatalf("Expected not equal to actual")
}
