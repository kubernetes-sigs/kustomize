// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package filters_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	. "sigs.k8s.io/kustomize/kyaml/kio/filters"
	"sigs.k8s.io/kustomize/kyaml/kio/filters/testyaml"
)

// TestFormatInput_configMap verifies a ConfigMap yaml is formatted correctly
func TestFormatInput_configMap(t *testing.T) {
	y := `


# this formatting is intentionally weird

apiVersion: v1
# this is data
data:
  # this is color
  color: purple
  # that was color

  # this is textmode
  textmode: "true"
  # this is how
  how: fairlyNice



kind: ConfigMap


metadata:
  selfLink: /api/v1/namespaces/default/configmaps/config-multi-env-files
  namespace: default
  creationTimestamp: 2017-12-27T18:38:34Z
  name: config-multi-env-files
  resourceVersion: "810136"
  uid: 252c4572-eb35-11e7-887b-42010a8002b8  # keep no trailing linefeed`

	expected := `# this formatting is intentionally weird

apiVersion: v1
kind: ConfigMap
metadata:
  name: config-multi-env-files
  namespace: default
  creationTimestamp: 2017-12-27T18:38:34Z
  resourceVersion: "810136"
  selfLink: /api/v1/namespaces/default/configmaps/config-multi-env-files
  uid: 252c4572-eb35-11e7-887b-42010a8002b8 # keep no trailing linefeed
# this is data
data:
  # this is color
  color: purple
  # that was color

  # this is how
  how: fairlyNice
  # this is textmode
  textmode: "true"
`

	s, err := FormatInput(strings.NewReader(y))
	assert.NoError(t, err)
	assert.Equal(t, expected, s.String())
}

// TestFormatInput_deployment verifies a Deployment yaml is formatted correctly
func TestFormatInput_deployment(t *testing.T) {
	y := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  selector:
    matchLabels:
      app: nginx
  replicas: 3
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      # this is a container
      - ports:
        # this is a port
        - containerPort: 80
        name: b-nginx
        image: nginx:1.7.9
      # this is another container
      - name: a-nginx
        image: nginx:1.7.9
        ports:
        - containerPort: 80
`
	expected := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - # this is another container
        name: a-nginx
        image: nginx:1.7.9
        ports:
        - containerPort: 80
      - name: b-nginx
        image: nginx:1.7.9
        # this is a container
        ports:
        - # this is a port
          containerPort: 80
`
	s, err := FormatInput(strings.NewReader(y))
	assert.NoError(t, err)
	assert.Equal(t, expected, s.String())
}

// TestFormatInput_service verifies a Service yaml is formatted correctly
func TestFormatInput_service(t *testing.T) {

	y := `
apiVersion: v1
kind: Service
metadata:
  name: my-service
spec:
  selector:
    app: MyApp
  ports:
    - protocol: TCP
      port: 80
      targetPort: 9376
`
	expected := `apiVersion: v1
kind: Service
metadata:
  name: my-service
spec:
  selector:
    app: MyApp
  ports:
  - protocol: TCP
    port: 80
    targetPort: 9376
`
	s, err := FormatInput(strings.NewReader(y))
	assert.NoError(t, err)
	assert.Equal(t, expected, s.String())
}

// TestFormatInput_service verifies a Service yaml is formatted correctly
func TestFormatInput_validatingWebhookConfiguration(t *testing.T) {

	y := `
apiVersion: admissionregistration.k8s.io/v1beta1
kind: ValidatingWebhookConfiguration
metadata:
  name: <name of this configuration object>
webhooks:
- name: <webhook name, e.g., pod-policy.example.io>
  rules:
  - apiGroups:
    - ""
    apiVersions:
    - v1
    operations:
      - UPDATE # this list is indented by 2
      - CREATE
      - CONNECT
    resources:
    - pods # this list is not indented by 2
    scope: "Namespaced"
  clientConfig:
    service:
      namespace: <namespace of the front-end service>
      name: <name of the front-end service>
    caBundle: <pem encoded ca cert that signs the server cert used by the webhook>
  admissionReviewVersions:
  - v1beta1
  timeoutSeconds: 1
`
	expected := `apiVersion: admissionregistration.k8s.io/v1beta1
kind: ValidatingWebhookConfiguration
metadata:
  name: <name of this configuration object>
webhooks:
- name: <webhook name, e.g., pod-policy.example.io>
  admissionReviewVersions:
  - v1beta1
  clientConfig:
    service:
      name: <name of the front-end service>
      namespace: <namespace of the front-end service>
    caBundle: <pem encoded ca cert that signs the server cert used by the webhook>
  rules:
  - resources:
    - pods # this list is not indented by 2
    apiGroups:
    - ""
    apiVersions:
    - v1
    operations:
    - CONNECT
    - CREATE
    - UPDATE # this list is indented by 2
    scope: Namespaced
  timeoutSeconds: 1
`
	s, err := FormatInput(strings.NewReader(y))
	assert.NoError(t, err)
	assert.Equal(t, expected, s.String())
}

// TestFormatInput_unKnownType verifies an unknown type yaml is formatted correctly
func TestFormatInput_unKnownType(t *testing.T) {
	y := `
spec:
  template:
    spec:
      # these shouldn't be sorted because the type isn't whitelisted
      containers:
      - name: b
      - name: a
  replicas: 1
status:
  conditions:
  - 3
  - 1
  - 2
other:
  b: a1
  a: b1
apiVersion: example.com/v1beta1
kind: MyType
`

	expected := `apiVersion: example.com/v1beta1
kind: MyType
spec:
  replicas: 1
  template:
    spec:
      # these shouldn't be sorted because the type isn't whitelisted
      containers:
      - name: b
      - name: a
status:
  conditions:
  - 3
  - 1
  - 2
other:
  a: b1
  b: a1
`
	s, err := FormatInput(strings.NewReader(y))
	assert.NoError(t, err)
	assert.Equal(t, expected, s.String())
}

// TestFormatInput_deployment verifies a Deployment yaml is formatted correctly
func TestFormatInput_resources(t *testing.T) {
	input := &bytes.Buffer{}
	_, err := io.Copy(input, bytes.NewReader(testyaml.UnformattedYaml1))
	assert.NoError(t, err)
	_, err = io.Copy(input, strings.NewReader("---\n"))
	assert.NoError(t, err)
	_, err = io.Copy(input, bytes.NewReader(testyaml.UnformattedYaml2))
	assert.NoError(t, err)

	expectedOutput := &bytes.Buffer{}
	_, err = io.Copy(expectedOutput, bytes.NewReader(testyaml.FormattedYaml1))
	assert.NoError(t, err)
	_, err = io.Copy(expectedOutput, strings.NewReader("---\n"))
	assert.NoError(t, err)
	_, err = io.Copy(expectedOutput, bytes.NewReader(testyaml.FormattedYaml2))
	assert.NoError(t, err)

	s, err := FormatInput(input)
	assert.NoError(t, err)
	assert.Equal(t, expectedOutput.String(), s.String())
}

//
func TestFormatInput_failMissingKind(t *testing.T) {
	y := `
spec:
  template:
    spec:
      containers:
      - b
      - a
  replicas: 1
status:
  conditions:
  - 3
  - 1
  - 2
other:
  b: a1
  a: b1
apiVersion: example.com/v1beta1
`

	b, err := FormatInput(strings.NewReader(y))
	assert.NoError(t, err)
	assert.Equal(t, strings.TrimLeft(y, "\n"), b.String())
}

func TestFormatInput_failMissingApiVersion(t *testing.T) {
	y := `
spec:
  template:
    spec:
      containers:
      - a
      - b
  replicas: 1
status:
  conditions:
  - 3
  - 1
  - 2
other:
  b: a1
  a: b1
kind: MyKind
`

	b, err := FormatInput(strings.NewReader(y))
	assert.NoError(t, err)
	assert.Equal(t, strings.TrimLeft(y, "\n"), b.String())
}

func TestFormatInput_failUnmarshal(t *testing.T) {
	y := `
spec:
  template:
    spec:
      containers:
      - a
      - b
  replicas: 1
status:
  conditions:
  - 3
  - 1
  - 2
other:
	b: a1
	a: b1
kind: MyKind
apiVersion: example.com/v1beta1
`

	_, err := FormatInput(strings.NewReader(y))
	assert.EqualError(t, err, "yaml: line 15: found character that cannot start any token")
}

// TestFormatFileOrDirectory_yamlExtFile verifies that FormatFileOrDirectory will format a file
// with a .yaml extension.
func TestFormatFileOrDirectory_yamlExtFile(t *testing.T) {
	// write the unformatted file
	f, err := ioutil.TempFile("", "yamlfmt*.yaml")
	if !assert.NoError(t, err) {
		return
	}
	defer os.Remove(f.Name())
	err = ioutil.WriteFile(f.Name(), testyaml.UnformattedYaml1, 0600)
	if !assert.NoError(t, err) {
		return
	}

	// format the file
	err = FormatFileOrDirectory(f.Name())
	if !assert.NoError(t, err) {
		return
	}

	// check the result is formatted
	b, err := ioutil.ReadFile(f.Name())
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, string(testyaml.FormattedYaml1), string(b))
}

func TestFormatFileOrDirectory_multipleYamlEntries(t *testing.T) {
	// write the unformatted file
	f, err := ioutil.TempFile("", "yamlfmt*.yaml")
	assert.NoError(t, err)
	defer os.Remove(f.Name())
	err = ioutil.WriteFile(f.Name(),
		[]byte(string(testyaml.UnformattedYaml1)+"---\n"+string(testyaml.UnformattedYaml2)), 0600)
	assert.NoError(t, err)

	// format the file
	err = FormatFileOrDirectory(f.Name())
	assert.NoError(t, err)

	// check the result is formatted
	b, err := ioutil.ReadFile(f.Name())
	assert.NoError(t, err)
	assert.Equal(t, string(testyaml.FormattedYaml1)+"---\n"+string(testyaml.FormattedYaml2), string(b))
}

// TestFormatFileOrDirectory_ymlExtFile verifies that FormatFileOrDirectory will format a file
// with a .yml extension.
func TestFormatFileOrDirectory_ymlExtFile(t *testing.T) {
	// write the unformatted file
	f, err := ioutil.TempFile("", "yamlfmt*.yml")
	assert.NoError(t, err)
	defer os.Remove(f.Name())
	err = ioutil.WriteFile(f.Name(), testyaml.UnformattedYaml1, 0600)
	assert.NoError(t, err)

	// format the file
	err = FormatFileOrDirectory(f.Name())
	assert.NoError(t, err)

	// check the result is formatted
	b, err := ioutil.ReadFile(f.Name())
	assert.NoError(t, err)
	assert.Equal(t, string(testyaml.FormattedYaml1), string(b))
}

// TestFormatFileOrDirectory_skipYamlExtFileWithJson verifies that the json content is formatted
// as yaml
func TestFormatFileOrDirectory_YamlExtFileWithJson(t *testing.T) {
	// write the unformatted JSON file contents
	f, err := ioutil.TempFile("", "yamlfmt*.yaml")
	assert.NoError(t, err)
	defer os.Remove(f.Name())
	err = ioutil.WriteFile(f.Name(), testyaml.UnformattedJson1, 0600)
	assert.NoError(t, err)

	// format the file
	err = FormatFileOrDirectory(f.Name())
	assert.NoError(t, err)

	// check the result is formatted as yaml
	b, err := ioutil.ReadFile(f.Name())
	assert.NoError(t, err)
	assert.Equal(t, string(testyaml.FormattedYaml1), string(b))
}

// TestFormatFileOrDirectory_partialKubernetesYamlFile verifies that if a yaml file contains both
// Kubernetes and non-Kubernetes documents, it will only format the Kubernetes documents
func TestFormatFileOrDirectory_partialKubernetesYamlFile(t *testing.T) {
	// write the unformatted file
	f, err := ioutil.TempFile("", "yamlfmt*.yaml")
	assert.NoError(t, err)
	defer os.Remove(f.Name())
	err = ioutil.WriteFile(f.Name(), []byte(string(testyaml.UnformattedYaml1)+`---
status:
  conditions:
  - 3
  - 1
  - 2
spec: a
---
`+string(testyaml.UnformattedYaml2)), 0600)
	assert.NoError(t, err)

	// format the file
	err = FormatFileOrDirectory(f.Name())
	assert.NoError(t, err)

	// check the result is  NOT formatted
	b, err := ioutil.ReadFile(f.Name())
	assert.NoError(t, err)
	assert.Equal(t, string(testyaml.FormattedYaml1)+`---
status:
  conditions:
  - 3
  - 1
  - 2
spec: a
---
`+string(testyaml.FormattedYaml2), string(b))
}

// TestFormatFileOrDirectory_nonKubernetesYamlFile verifies that if a yaml file does not contain
// kubernetes
func TestFormatFileOrDirectory_skipNonKubernetesYamlFile(t *testing.T) {
	// write the unformatted JSON file contents
	f, err := ioutil.TempFile("", "yamlfmt*.yaml")
	assert.NoError(t, err)
	defer os.Remove(f.Name())
	err = ioutil.WriteFile(f.Name(), []byte(`
status:
  conditions:
  - 3
  - 1
  - 2
spec: a
`), 0600)
	assert.NoError(t, err)

	// format the file
	err = FormatFileOrDirectory(f.Name())
	assert.NoError(t, err)

	// check the result is formatted as yaml
	b, err := ioutil.ReadFile(f.Name())
	assert.NoError(t, err)
	assert.Equal(t, `status:
  conditions:
  - 3
  - 1
  - 2
spec: a
`, string(b))
}

// TestFormatFileOrDirectory_jsonFile should not fmt the file even though it contains yaml.
func TestFormatFileOrDirectory_skipJsonExtFile(t *testing.T) {
	f, err := ioutil.TempFile("", "yamlfmt*.json")
	assert.NoError(t, err)
	defer os.Remove(f.Name())
	err = ioutil.WriteFile(f.Name(), testyaml.UnformattedYaml1, 0600)
	assert.NoError(t, err)

	err = FormatFileOrDirectory(f.Name())
	assert.NoError(t, err)

	b, err := ioutil.ReadFile(f.Name())
	assert.NoError(t, err)

	assert.Equal(t, string(testyaml.UnformattedYaml1), string(b))
}

// TestFormatFileOrDirectory_directory verifies that yaml files will be formatted,
// and other files will be ignored
func TestFormatFileOrDirectory_directory(t *testing.T) {
	d, err := ioutil.TempDir("", "yamlfmt")
	assert.NoError(t, err)

	err = os.Mkdir(filepath.Join(d, "config"), 0700)
	assert.NoError(t, err)

	err = ioutil.WriteFile(filepath.Join(d, "c1.yaml"), testyaml.UnformattedYaml1, 0600)
	assert.NoError(t, err)

	err = ioutil.WriteFile(filepath.Join(d, "config", "c2.yaml"), testyaml.UnformattedYaml2, 0600)
	assert.NoError(t, err)

	err = ioutil.WriteFile(filepath.Join(d, "README.md"), []byte(`# Markdown`), 0600)
	assert.NoError(t, err)

	err = FormatFileOrDirectory(d)
	assert.NoError(t, err)

	b, err := ioutil.ReadFile(filepath.Join(d, "c1.yaml"))
	assert.NoError(t, err)
	assert.Equal(t, string(testyaml.FormattedYaml1), string(b))

	b, err = ioutil.ReadFile(filepath.Join(d, "config", "c2.yaml"))
	assert.NoError(t, err)
	assert.Equal(t, string(testyaml.FormattedYaml2), string(b))

	b, err = ioutil.ReadFile(filepath.Join(d, "README.md"))
	assert.NoError(t, err)
	assert.Equal(t, `# Markdown`, string(b))

	// verify no additional files were created
	files := []string{
		".", "c1.yaml", "README.md", "config", filepath.Join("config", "c2.yaml")}
	err = filepath.Walk(d, func(path string, info os.FileInfo, err error) error {
		assert.NoError(t, err)
		path, err = filepath.Rel(d, path)
		assert.NoError(t, err)
		assert.Contains(t, files, path)
		return nil
	})
	assert.NoError(t, err)
}

// TestFormatFileOrDirectory_trimWhiteSpace verifies that trailling and leading whitespace is
// trimmed
func TestFormatFileOrDirectory_trimWhiteSpace(t *testing.T) {
	f, err := ioutil.TempFile("", "yamlfmt*.yaml")
	assert.NoError(t, err)
	defer os.Remove(f.Name())
	err = ioutil.WriteFile(f.Name(), []byte("\n\n"+string(testyaml.UnformattedYaml1)+"\n\n"), 0600)
	assert.NoError(t, err)

	err = FormatFileOrDirectory(f.Name())
	assert.NoError(t, err)

	b, err := ioutil.ReadFile(f.Name())
	assert.NoError(t, err)

	assert.Equal(t, string(testyaml.FormattedYaml1), string(b))
}
