// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package comments

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestCopyComments(t *testing.T) {
	from, err := yaml.Parse(`# A
#
# B

# C
apiVersion: apps/v1
kind: Deployment
spec: # comment 1
  # comment 2
  replicas: 3 # comment 3
  # comment 4
`)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	to, err := yaml.Parse(`apiVersion: apps/v1
kind: Deployment
spec:
  replicas: 4
`)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	err = CopyComments(from, to)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	actual, err := to.String()
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	expected := `# A
#
# B

# C
apiVersion: apps/v1
kind: Deployment
spec: # comment 1
  # comment 2
  replicas: 4 # comment 3
  # comment 4
`

	if !assert.Equal(t, expected, actual) {
		t.FailNow()
	}
}
