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

package configmapandsecret

import (
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/types"
)

func TestMakeSecretNoCommands(t *testing.T) {
	factory := NewSecretFactory(fs.MakeFakeFS(), "/")
	args := types.SecretArgs{
		GeneratorArgs: types.GeneratorArgs{Name: "apple"},
		Type:          "Opaque",
		CommandSources: types.CommandSources{
			Commands:   nil,
			EnvCommand: "",
		}}
	s, err := factory.MakeSecret(&args, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.ObjectMeta.Name != "apple" {
		t.Fatalf("unexpected name: %v", s.ObjectMeta.Name)
	}
	if len(s.Data) > 0 || len(s.StringData) > 0 {
		t.Fatalf("unexpected data: %v", s)
	}
}

func TestMakeSecretNoCommandsBadDir(t *testing.T) {
	factory := NewSecretFactory(fs.MakeFakeFS(), "/does/not/exist")
	args := types.SecretArgs{
		GeneratorArgs: types.GeneratorArgs{Name: "envConfigMap"},
		Type:          "Opaque",
		CommandSources: types.CommandSources{
			Commands:   nil,
			EnvCommand: "",
		}}
	_, err := factory.MakeSecret(&args, nil)
	if err == nil {
		t.Fatalf("expected error: %v", err)
	}
	if !strings.Contains(err.Error(), "not a directory") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMakeSecretEmptyCommandMap(t *testing.T) {
	factory := NewSecretFactory(fs.MakeFakeFS(), "/")
	args := types.SecretArgs{
		GeneratorArgs: types.GeneratorArgs{Name: "envConfigMap"},
		Type:          "Opaque",
		CommandSources: types.CommandSources{
			// TODO try: map[string]string{"commandName": "bogusCommand bogusArg"},
			Commands:   nil,
			EnvCommand: "echo beans",
		}}
	s, err := factory.MakeSecret(&args, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatalf("nil result")
	}
	v, ok := s.Data["beans"]
	if !ok {
		t.Fatalf("expected beans")
	}
	if len(v) > 0 {
		t.Fatalf("unexpected data")
	}
}

func TestMakeSecretWithCommandMap(t *testing.T) {
	factory := NewSecretFactory(fs.MakeFakeFS(), "/")
	args := types.SecretArgs{
		GeneratorArgs: types.GeneratorArgs{Name: "envConfigMap"},
		Type:          "Opaque",
		CommandSources: types.CommandSources{
			Commands: map[string]string{"commandName": "echo beans"},
		}}
	s, err := factory.MakeSecret(&args, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatalf("nil result")
	}
	v, ok := s.Data["commandName"]
	if !ok {
		t.Fatalf("expected something for commandName")
	}
	if string(v) != "beans\n" {
		t.Fatalf("unexpected data: %s", string(v))
	}
}
