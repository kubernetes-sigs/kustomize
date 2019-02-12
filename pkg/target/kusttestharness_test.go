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

package target_test

// A collection of utilities used in target tests.

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/pkg/constants"
	"sigs.k8s.io/kustomize/pkg/internal/loadertest"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	. "sigs.k8s.io/kustomize/pkg/target"
	"sigs.k8s.io/kustomize/pkg/transformers/config/defaultconfig"
	"sigs.k8s.io/kustomize/pkg/types"
)

type KustTestHarness struct {
	t   *testing.T
	rf  *resmap.Factory
	ldr loadertest.FakeLoader
}

func NewKustTestHarness(t *testing.T, path string) *KustTestHarness {
	return &KustTestHarness{
		t: t,
		rf: resmap.NewFactory(resource.NewFactory(
			kunstruct.NewKunstructuredFactoryImpl())),
		ldr: loadertest.NewFakeLoader(path)}
}

func (th *KustTestHarness) makeKustTarget() *KustTarget {
	kt, err := NewKustTarget(
		th.ldr, th.rf, transformer.NewFactoryImpl())
	if err != nil {
		th.t.Fatalf("Unexpected construction error %v", err)
	}
	return kt
}

func (th *KustTestHarness) writeF(dir string, content string) {
	err := th.ldr.AddFile(dir, []byte(content))
	if err != nil {
		th.t.Fatalf("failed write to %s; %v", dir, err)
	}
}

func (th *KustTestHarness) writeK(dir string, content string) {
	th.writeF(filepath.Join(dir, constants.KustomizationFileNames[0]), `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`+content)
}

func (th *KustTestHarness) fromMap(m map[string]interface{}) *resource.Resource {
	return th.rf.RF().FromMap(m)
}

func (th *KustTestHarness) fromMapAndOption(m map[string]interface{}, args *types.GeneratorArgs, option *types.GeneratorOptions) *resource.Resource {
	return th.rf.RF().FromMapAndOption(m, args, option)
}

func (th *KustTestHarness) writeDefaultConfigs(fName string) {
	m := defaultconfig.GetDefaultFieldSpecsAsMap()
	var content []byte
	for _, tCfg := range m {
		content = append(content, []byte(tCfg)...)
	}
	err := th.ldr.AddFile(fName, content)
	if err != nil {
		th.t.Fatalf("unable to add file %s", fName)
	}
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

func (th *KustTestHarness) assertActualEqualsExpected(
	m resmap.ResMap, expected string) {
	if m == nil {
		th.t.Fatalf("Map should not be nil.")
	}
	// Ignore leading linefeed in expected value
	// to ease readability of tests.
	if len(expected) > 0 && expected[0] == 10 {
		expected = expected[1:]
	}
	actual, err := m.EncodeAsYaml()
	if err != nil {
		th.t.Fatalf("Unexpected err: %v", err)
	}
	if string(actual) != expected {
		th.reportDiffAndFail(actual, expected)
	}
}

// Pretty printing of file differences.
func (th *KustTestHarness) reportDiffAndFail(actual []byte, expected string) {
	sE, maxLen := convertToArray(expected)
	sA, _ := convertToArray(string(actual))
	fmt.Println("===== ACTUAL BEGIN ========================================")
	fmt.Print(string(actual))
	fmt.Println("===== ACTUAL END ==========================================")
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
	th.t.Fatalf("Expected not equal to actual")
}
