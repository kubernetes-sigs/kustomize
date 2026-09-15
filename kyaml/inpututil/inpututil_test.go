// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package inpututil_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/inpututil"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestWrapErrorWithFileReturnsNilForNilError(t *testing.T) {
	assert.Nil(t, inpututil.WrapErrorWithFile(nil, yaml.ResourceMeta{}))
}

func TestWrapErrorWithFileUsesCurrentAnnotations(t *testing.T) {
	meta := yaml.ResourceMeta{
		ObjectMeta: yaml.ObjectMeta{
			Annotations: map[string]string{
				kioutil.PathAnnotation:  "resources/foo.yaml",
				kioutil.IndexAnnotation: "2",
			},
		},
	}
	err := inpututil.WrapErrorWithFile(errors.New("boom"), meta)
	assert.ErrorContains(t, err, "resources/foo.yaml")
	assert.ErrorContains(t, err, "2")
}

func TestWrapErrorWithFileFallsBackToLegacyAnnotations(t *testing.T) {
	meta := yaml.ResourceMeta{
		ObjectMeta: yaml.ObjectMeta{
			Annotations: map[string]string{
				kioutil.LegacyPathAnnotation:  "resources/bar.yaml",
				kioutil.LegacyIndexAnnotation: "5",
			},
		},
	}
	err := inpututil.WrapErrorWithFile(errors.New("boom"), meta)
	assert.ErrorContains(t, err, "resources/bar.yaml")
	assert.ErrorContains(t, err, "5")
}
