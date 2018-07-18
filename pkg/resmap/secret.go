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

package resmap

import (
	"context"
	"os/exec"
	"path/filepath"
	"time"

	"os"

	"github.com/kubernetes-sigs/kustomize/pkg/resource"
	"github.com/kubernetes-sigs/kustomize/pkg/types"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

func newResourceFromSecretGenerator(p string, sArgs types.SecretArgs) (*resource.Resource, error) {
	s, err := makeSecret(p, sArgs)
	if err != nil {
		return nil, errors.Wrap(err, "makeSecret")
	}
	if sArgs.Behavior == "" {
		sArgs.Behavior = "create"
	}
	return resource.NewResourceWithBehavior(
		s, resource.NewGenerationBehavior(sArgs.Behavior))
}

func makeSecret(p string, sArgs types.SecretArgs) (*corev1.Secret, error) {
	s := &corev1.Secret{}
	s.APIVersion = "v1"
	s.Kind = "Secret"
	s.Name = sArgs.Name
	s.Type = corev1.SecretType(sArgs.Type)
	if s.Type == "" {
		s.Type = corev1.SecretTypeOpaque
	}
	s.Data = map[string][]byte{}

	for k, v := range sArgs.Commands {
		out, err := createSecretKey(p, v)
		if err != nil {
			return nil, errors.Wrap(err, "createSecretKey")
		}
		s.Data[k] = out
	}
	return s, nil
}

// Run a command, return its output as the secret.
func createSecretKey(wd string, command string) ([]byte, error) {
	fi, err := os.Stat(wd)
	if err != nil || !fi.IsDir() {
		wd = filepath.Dir(wd)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = wd
	return cmd.Output()
}

// NewResMapFromSecretArgs takes a SecretArgs slice and executes its command in directory p
// then writes the output to a Resource slice and return it.
func NewResMapFromSecretArgs(p string, secretList []types.SecretArgs) (ResMap, error) {
	allResources := []*resource.Resource{}
	for _, secret := range secretList {
		res, err := newResourceFromSecretGenerator(p, secret)
		if err != nil {
			return nil, errors.Wrap(err, "newResourceFromSecretGenerator")
		}
		allResources = append(allResources, res)
	}
	return newResMapFromResourceSlice(allResources)
}
