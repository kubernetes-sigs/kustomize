/*
Copyright 2019 The Kubernetes Authors.

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

package add

import (
	"testing"

	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/types"
)

func TestNewCmdAddSecretIsNotNil(t *testing.T) {
	if newCmdAddSecret(fs.MakeFakeFS(), nil) == nil {
		t.Fatal("newCmdAddSecret shouldn't be nil")
	}
}

func TestMakeSecretArgs(t *testing.T) {
	secretName := "test-secret-name"

	kustomization := &types.Kustomization{
		NamePrefix: "test-name-prefix",
	}

	secretType := "Opaque"

	if len(kustomization.SecretGenerator) != 0 {
		t.Fatal("Initial kustomization should not have any secrets")
	}
	args := makeSecretArgs(kustomization, secretName, secretType)

	if args == nil {
		t.Fatalf("args should always be non-nil")
	}

	if len(kustomization.SecretGenerator) != 1 {
		t.Fatalf("Kustomization should have newly created secret")
	}

	if &kustomization.SecretGenerator[len(kustomization.SecretGenerator)-1] != args {
		t.Fatalf("Pointer address for newly inserted secret generator should be same")
	}

	args2 := makeSecretArgs(kustomization, secretName, secretType)

	if args2 != args {
		t.Fatalf("should have returned an existing args with name: %v", secretName)
	}

	if len(kustomization.SecretGenerator) != 1 {
		t.Fatalf("Should not insert secret for an existing name: %v", secretName)
	}
}

func TestMergeFlagsIntoSecretArgs_LiteralSources(t *testing.T) {
	var kv []types.KVSource

	mergeFlagsIntoSecretArgs(&kv, flagsAndArgs{LiteralSources: []string{"k1=v1"}})

	if len(kv) != 1 {
		t.Fatalf("Initial literal source should have been added")
	}

	mergeFlagsIntoSecretArgs(&kv, flagsAndArgs{LiteralSources: []string{"k2=v2"}})

	if len(kv) != 2 {
		t.Fatalf("Second literal source should have been added")
	}
}

func TestMergeFlagsIntoSecretArgs_FileSources(t *testing.T) {
	var kv []types.KVSource

	mergeFlagsIntoSecretArgs(&kv, flagsAndArgs{FileSources: []string{"file1"}})

	if len(kv) != 1 {
		t.Fatalf("Initial file source should have been added")
	}

	mergeFlagsIntoSecretArgs(&kv, flagsAndArgs{FileSources: []string{"file2"}})

	if len(kv) != 2 {
		t.Fatalf("Second file source should have been added")
	}
}

func TestMergeFlagsIntoSecretArgs_EnvSource(t *testing.T) {
	envFileName := "env1"
	envFileName2 := "env2"
	var kv []types.KVSource

	mergeFlagsIntoSecretArgs(&kv, flagsAndArgs{EnvFileSource: envFileName})

	if len(kv) != 1 {
		t.Fatalf("Initial env source should have been added")
	}

	mergeFlagsIntoSecretArgs(&kv, flagsAndArgs{EnvFileSource: envFileName2})
	if len(kv) != 2 {
		t.Fatalf("Second env source should have been added")
	}
}
