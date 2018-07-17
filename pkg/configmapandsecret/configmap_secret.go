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

// Package configmapandsecret generates configmaps and secrets per generator rules.
package configmapandsecret

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	cutil "github.com/kubernetes-sigs/kustomize/pkg/configmapandsecret/util"
	"github.com/kubernetes-sigs/kustomize/pkg/fs"
	"github.com/kubernetes-sigs/kustomize/pkg/hash"
	"github.com/kubernetes-sigs/kustomize/pkg/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// MakeConfigmapAndGenerateName makes a configmap and returns the configmap and the name appended with a hash.
func MakeConfigmapAndGenerateName(fs fs.FileSystem, cm types.ConfigMapArgs) (*unstructured.Unstructured, string, error) {
	corev1CM, err := makeConfigMap(fs, cm)
	if err != nil {
		return nil, "", err
	}
	h, err := hash.ConfigMapHash(corev1CM)
	if err != nil {
		return nil, "", err
	}
	nameWithHash := fmt.Sprintf("%s-%s", corev1CM.GetName(), h)
	unstructuredCM, err := objectToUnstructured(corev1CM)
	return unstructuredCM, nameWithHash, err
}

// MakeSecretAndGenerateName returns a secret with the name appended with a hash.
func MakeSecretAndGenerateName(fs fs.FileSystem, secret types.SecretArgs, path string) (*unstructured.Unstructured, string, error) {
	corev1Secret, err := makeSecret(fs, secret, path)
	if err != nil {
		return nil, "", err
	}
	h, err := hash.SecretHash(corev1Secret)
	if err != nil {
		return nil, "", err
	}
	nameWithHash := fmt.Sprintf("%s-%s", secret.Name, h)
	unstructuredCM, err := objectToUnstructured(corev1Secret)
	return unstructuredCM, nameWithHash, err
}

func objectToUnstructured(in runtime.Object) (*unstructured.Unstructured, error) {
	marshaled, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	var out unstructured.Unstructured
	err = out.UnmarshalJSON(marshaled)
	return &out, err
}

func makeConfigMap(fs fs.FileSystem, cm types.ConfigMapArgs) (*corev1.ConfigMap, error) {
	corev1cm := &corev1.ConfigMap{}
	corev1cm.APIVersion = "v1"
	corev1cm.Kind = "ConfigMap"
	corev1cm.Name = cm.Name
	corev1cm.Data = map[string]string{}

	if cm.EnvSource != "" {
		if err := cutil.HandleConfigMapFromEnvFileSource(fs, corev1cm, cm.EnvSource); err != nil {
			return nil, err
		}
	}
	if cm.FileSources != nil {
		if err := cutil.HandleConfigMapFromFileSources(fs, corev1cm, cm.FileSources); err != nil {
			return nil, err
		}
	}
	if cm.LiteralSources != nil {
		if err := cutil.HandleConfigMapFromLiteralSources(corev1cm, cm.LiteralSources); err != nil {
			return nil, err
		}
	}

	return corev1cm, nil
}

func makeSecret(fs fs.FileSystem, secret types.SecretArgs, path string) (*corev1.Secret, error) {
	corev1secret := &corev1.Secret{}
	corev1secret.APIVersion = "v1"
	corev1secret.Kind = "Secret"
	corev1secret.Name = secret.Name
	corev1secret.Type = corev1.SecretType(secret.Type)
	if corev1secret.Type == "" {
		corev1secret.Type = corev1.SecretTypeOpaque
	}
	corev1secret.Data = map[string][]byte{}

	for k, v := range secret.Commands {
		out, err := createSecretKey(fs, path, v)
		if err != nil {
			return nil, err
		}
		corev1secret.Data[k] = out
	}

	return corev1secret, nil
}

func createSecretKey(fs fs.FileSystem, wd string, command string) ([]byte, error) {
	fi, err := fs.Stat(wd)
	if err != nil {
		switch err := err.(type) {
		case *os.PathError:
			return nil, fmt.Errorf("unable to get info %s: %v", wd, err.Err)
		default:
			return nil, fmt.Errorf("unable to get info %s: %v", wd, err)
		}
	}
	if err != nil || !fi.IsDir() {
		wd = filepath.Dir(wd)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = wd

	return cmd.Output()
}
