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

package resource

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	manifest "k8s.io/kubectl/pkg/apis/manifest/v1alpha1"
)

func newFromSecretGenerator(p string, s manifest.SecretArgs) (*Resource, error) {
	corev1secret := &corev1.Secret{}
	corev1secret.APIVersion = "v1"
	corev1secret.Kind = "Secret"
	corev1secret.Name = s.Name
	corev1secret.Type = corev1.SecretType(s.Type)
	if corev1secret.Type == "" {
		corev1secret.Type = corev1.SecretTypeOpaque
	}
	corev1secret.Data = map[string][]byte{}

	for k, v := range s.Commands {
		out, err := createSecretKey(p, v)
		if err != nil {
			return nil, err
		}
		corev1secret.Data[k] = out
	}

	obj, err := objectToUnstructured(corev1secret)

	if err != nil {
		return nil, err
	}

	return &Resource{Data: obj, Behavior: s.Behavior}, nil
}

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

// NewFromSecretGenerators takes a SecretGenerator slice and executes its command in directory p
// then writes the output to a Resource slice and return it.
func NewFromSecretGenerators(p string, secretList []manifest.SecretArgs) (ResourceCollection, error) {
	allResources := []*Resource{}
	for _, secret := range secretList {
		res, err := newFromSecretGenerator(p, secret)
		if err != nil {
			return nil, err
		}
		allResources = append(allResources, res)
	}
	return resourceCollectionFromResources(allResources)
}
