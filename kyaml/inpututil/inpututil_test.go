// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package inpututil_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/inpututil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestMapInputsEWrapsErrorWithFileAndIndex(t *testing.T) {
	testCases := map[string]struct {
		annotations string
	}{
		"internal annotations": {annotations: `
      internal.config.kubernetes.io/path: 'deployment.yaml'
      internal.config.kubernetes.io/index: '3'`},
		"legacy annotations": {annotations: `
      config.kubernetes.io/path: 'deployment.yaml'
      config.kubernetes.io/index: '3'`},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			in, err := yaml.Parse(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  annotations:` + tc.annotations + `
`)
			require.NoError(t, err)

			err = inpututil.MapInputsE([]*yaml.RNode{in},
				func(*yaml.RNode, yaml.ResourceMeta) error {
					return errors.New("boom")
				})
			require.EqualError(t, err, "deployment.yaml [3]: boom")
		})
	}
}
