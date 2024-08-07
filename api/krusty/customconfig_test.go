// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func makeBaseReferencingCustomConfig(th kusttest_test.Harness) {
	th.WriteK("base", `
namePrefix: x-
commonLabels:
  app: myApp
vars:
- name: APRIL_DIET
  objref:
    group: foo
    version: v1
    kind: Giraffe
    name: april
  fieldref:
    fieldpath: spec.diet
- name: KOKO_DIET
  objref:
    group: foo
    version: v1
    kind: Gorilla
    name: koko
  fieldref:
    fieldpath: spec.diet
resources:
- animalPark.yaml
- giraffes.yaml
- gorilla.yaml
configurations:
- config/defaults.yaml
- config/custom.yaml
`)
	th.WriteF("base/giraffes.yaml", `
apiVersion: foo/v1
kind: Giraffe
metadata:
  name: april
spec:
  diet: mimosa
  location: NE
---
apiVersion: foo/v1
kind: Giraffe
metadata:
  name: may
spec:
  diet: acacia
  location: SE
`)
	th.WriteF("base/gorilla.yaml", `
apiVersion: foo/v1
kind: Gorilla
metadata:
  name: koko
spec:
  diet: bambooshoots
  location: SW
`)
	th.WriteF("base/animalPark.yaml", `
apiVersion: foo/v1
kind: AnimalPark
metadata:
  name: sandiego
spec:
  gorillaRef:
    name: koko
  giraffeRef:
    name: april
  food:
  - "$(APRIL_DIET)"
  - "$(KOKO_DIET)"
`)
}

func TestCustomConfig(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	makeBaseReferencingCustomConfig(th)
	th.WriteLegacyConfigs("base/config/defaults.yaml")
	th.WriteF("base/config/custom.yaml", `
nameReference:
- kind: Gorilla
  fieldSpecs:
  - kind: AnimalPark
    path: spec/gorillaRef/name
- kind: Giraffe
  fieldSpecs:
  - kind: AnimalPark
    path: spec/giraffeRef/name
varReference:
- path: spec/food
  kind: AnimalPark
`)
	m := th.Run("base", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: foo/v1
kind: AnimalPark
metadata:
  labels:
    app: myApp
  name: x-sandiego
spec:
  food:
  - mimosa
  - bambooshoots
  giraffeRef:
    name: x-april
  gorillaRef:
    name: x-koko
---
apiVersion: foo/v1
kind: Giraffe
metadata:
  labels:
    app: myApp
  name: x-april
spec:
  diet: mimosa
  location: NE
---
apiVersion: foo/v1
kind: Giraffe
metadata:
  labels:
    app: myApp
  name: x-may
spec:
  diet: acacia
  location: SE
---
apiVersion: foo/v1
kind: Gorilla
metadata:
  labels:
    app: myApp
  name: x-koko
spec:
  diet: bambooshoots
  location: SW
`)
}

func TestCustomConfigWithDefaultOverspecification(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	makeBaseReferencingCustomConfig(th)
	th.WriteLegacyConfigs("base/config/defaults.yaml")
	// Specifying namePrefix here conflicts with (is the same as)
	// the defaults written above.  This is intentional in the
	// test to assure duplicate config doesn't cause problems.
	th.WriteF("base/config/custom.yaml", `
namePrefix:
- path: metadata/name
nameReference:
- kind: Gorilla
  fieldSpecs:
  - kind: AnimalPark
    path: spec/gorillaRef/name
- kind: Giraffe
  fieldSpecs:
  - kind: AnimalPark
    path: spec/giraffeRef/name
varReference:
- path: spec/food
  kind: AnimalPark
`)
	m := th.Run("base", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: foo/v1
kind: AnimalPark
metadata:
  labels:
    app: myApp
  name: x-sandiego
spec:
  food:
  - mimosa
  - bambooshoots
  giraffeRef:
    name: x-april
  gorillaRef:
    name: x-koko
---
apiVersion: foo/v1
kind: Giraffe
metadata:
  labels:
    app: myApp
  name: x-april
spec:
  diet: mimosa
  location: NE
---
apiVersion: foo/v1
kind: Giraffe
metadata:
  labels:
    app: myApp
  name: x-may
spec:
  diet: acacia
  location: SE
---
apiVersion: foo/v1
kind: Gorilla
metadata:
  labels:
    app: myApp
  name: x-koko
spec:
  diet: bambooshoots
  location: SW
`)
}

func TestFixedBug605_BaseCustomizationAvailableInOverlay(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	makeBaseReferencingCustomConfig(th)
	th.WriteLegacyConfigs("base/config/defaults.yaml")
	th.WriteF("base/config/custom.yaml", `
nameReference:
- kind: Gorilla
  fieldSpecs:
  - group: foo
    version: v1
    kind: AnimalPark
    path: spec/gorillaRef/name
- kind: Giraffe
  fieldSpecs:
  - group: foo
    version: v1
    kind: AnimalPark
    path: spec/giraffeRef/name
varReference:
- path: spec/food
  group: foo
  version: v1
  kind: AnimalPark
`)
	th.WriteK("overlay", `
namePrefix: o-
commonLabels:
  movie: planetOfTheApes
patchesStrategicMerge:
- animalPark.yaml
resources:
- ../base
- ursus.yaml
`)
	th.WriteF("overlay/ursus.yaml", `
apiVersion: foo/v1
kind: Gorilla
metadata:
  name: ursus
spec:
  diet: heston
  location: Arizona
`)
	// The following replaces the gorillaRef in the AnimalPark.
	th.WriteF("overlay/animalPark.yaml", `
apiVersion: foo/v1
kind: AnimalPark
metadata:
  name: sandiego
spec:
  gorillaRef:
    name: ursus
`)
	m := th.Run("overlay", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: foo/v1
kind: AnimalPark
metadata:
  labels:
    app: myApp
    movie: planetOfTheApes
  name: o-x-sandiego
spec:
  food:
  - mimosa
  - bambooshoots
  giraffeRef:
    name: o-x-april
  gorillaRef:
    name: o-ursus
---
apiVersion: foo/v1
kind: Giraffe
metadata:
  labels:
    app: myApp
    movie: planetOfTheApes
  name: o-x-april
spec:
  diet: mimosa
  location: NE
---
apiVersion: foo/v1
kind: Giraffe
metadata:
  labels:
    app: myApp
    movie: planetOfTheApes
  name: o-x-may
spec:
  diet: acacia
  location: SE
---
apiVersion: foo/v1
kind: Gorilla
metadata:
  labels:
    app: myApp
    movie: planetOfTheApes
  name: o-x-koko
spec:
  diet: bambooshoots
  location: SW
---
apiVersion: foo/v1
kind: Gorilla
metadata:
  labels:
    movie: planetOfTheApes
  name: o-ursus
spec:
  diet: heston
  location: Arizona
`)
}

func TestLabelTransformerConfig(t *testing.T) {
	testCases := []struct {
		name              string
		kustomization     string
		transformerConfig string
		expectedResult    string
	}{
		{
			name: "includeSelectors=false, includeTemplates=false, include template via transformerConfig",
			kustomization: `configurations:
  - config/configurations.yaml

labels:
  - includeSelectors: false
    includeTemplates: false
    pairs:
      location: planet-earth
      environment: dev

resources:
  - resources/deployment.yaml
`,
			transformerConfig: `labels:
  - path: spec/template/metadata/labels
    create: true
    kind: Deployment
`,
			expectedResult: `apiVersion: apps/v1
kind: Deployment
metadata:
  creationTimestamp: null
  labels:
    app: sample-deploy
    environment: dev
    location: planet-earth
  name: sample-deploy
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sample-deploy
  strategy: {}
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: sample-deploy
        environment: dev
        location: planet-earth
    spec:
      containers:
      - image: hello-world:latest
        name: hello-world
`,
		},
		{
			name: "includeSelectors=true, includeTemplates=false, include template via transformerConfig",
			kustomization: `configurations:
  - config/configurations.yaml

labels:
  - includeSelectors: true
    includeTemplates: false
    pairs:
      location: planet-earth
      environment: dev

resources:
  - resources/deployment.yaml
`,
			transformerConfig: `labels:
  - path: spec/template/metadata/labels
    create: true
    kind: Deployment
`,
			expectedResult: `apiVersion: apps/v1
kind: Deployment
metadata:
  creationTimestamp: null
  labels:
    app: sample-deploy
    environment: dev
    location: planet-earth
  name: sample-deploy
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sample-deploy
      environment: dev
      location: planet-earth
  strategy: {}
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: sample-deploy
        environment: dev
        location: planet-earth
    spec:
      containers:
      - image: hello-world:latest
        name: hello-world
`,
		},
		{
			name: "includeSelectors=false, includeTemplates=true, no transformerConfig",
			kustomization: `labels:
  - includeSelectors: false
    includeTemplates: true
    pairs:
      location: planet-earth
      environment: dev

resources:
  - resources/deployment.yaml
`,
			expectedResult: `apiVersion: apps/v1
kind: Deployment
metadata:
  creationTimestamp: null
  labels:
    app: sample-deploy
    environment: dev
    location: planet-earth
  name: sample-deploy
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sample-deploy
  strategy: {}
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: sample-deploy
        environment: dev
        location: planet-earth
    spec:
      containers:
      - image: hello-world:latest
        name: hello-world
`,
		},
		{
			name: "includeSelectors=false, includeTemplates=false, no transformerConfig",
			kustomization: `labels:
  - includeSelectors: false
    includeTemplates: false
    pairs:
      location: planet-earth
      environment: dev

resources:
  - resources/deployment.yaml
`,
			expectedResult: `apiVersion: apps/v1
kind: Deployment
metadata:
  creationTimestamp: null
  labels:
    app: sample-deploy
    environment: dev
    location: planet-earth
  name: sample-deploy
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sample-deploy
  strategy: {}
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: sample-deploy
    spec:
      containers:
      - image: hello-world:latest
        name: hello-world
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			th := kusttest_test.MakeHarness(t)
			th.WriteK(".", tc.kustomization)
			th.WriteF("resources/deployment.yaml",
				`apiVersion: apps/v1
kind: Deployment
metadata:
  creationTimestamp: null
  labels:
    app: sample-deploy
  name: sample-deploy
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sample-deploy

  strategy: {}
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: sample-deploy
    spec:
      containers:
      - image: hello-world:latest
        name: hello-world
`)
			if tc.transformerConfig != "" {
				th.WriteF("config/configurations.yaml", tc.transformerConfig)
			}

			output := th.Run(".", th.MakeDefaultOptions())

			th.AssertActualEqualsExpected(output, tc.expectedResult)
		})
	}
}

func TestLabelTransformerConfigWithCustomResources(t *testing.T) {
	testCases := []struct {
		name              string
		kustomization     string
		transformerConfig string
		expectedResult    string
	}{
		{
			name: "include template via transformerConfig",
			kustomization: `configurations:
  - config/configurations.yaml

labels:
  - includeSelectors: false
    includeTemplates: false
    pairs:
      location: planet-earth
      environment: dev

resources:
  - resources/custom-resource.yaml
`,
			transformerConfig: `labels:
  - path: spec/template/metadata/labels
    create: true
    kind: SampleResource
`,
			expectedResult: `apiVersion: custom.example.org/v1
kind: SampleResource
metadata:
  labels:
    environment: dev
    location: planet-earth
  name: sample-resource
  namespace: sample-namespace
spec:
  template:
    metadata:
      labels:
        environment: dev
        location: planet-earth
    spec:
      containers:
      - env:
        - name: VARIABLE
          value: value
        image: index.docker.io/library/hello-world
`,
		},
		{
			name: "include selector via transformerConfig",
			kustomization: `configurations:
  - config/configurations.yaml

labels:
  - includeSelectors: false
    includeTemplates: false
    pairs:
      location: planet-earth
      environment: dev

resources:
  - resources/custom-resource.yaml
`,
			transformerConfig: `labels:
  - path: spec/selectors/labels
    create: true
    kind: SampleResource
`,
			expectedResult: `apiVersion: custom.example.org/v1
kind: SampleResource
metadata:
  labels:
    environment: dev
    location: planet-earth
  name: sample-resource
  namespace: sample-namespace
spec:
  selectors:
    labels:
      environment: dev
      location: planet-earth
  template:
    spec:
      containers:
      - env:
        - name: VARIABLE
          value: value
        image: index.docker.io/library/hello-world
`,
		},
		{
			name: "include selectors and labels via transformerConfig",
			kustomization: `configurations:
  - config/configurations.yaml

labels:
  - includeSelectors: false
    includeTemplates: false
    pairs:
      location: planet-earth
      environment: dev

resources:
  - resources/custom-resource.yaml
`,
			transformerConfig: `
labels:
  - path: spec/selectors/labels
    create: true
    kind: SampleResource
  - path: spec/template/metadata/labels
    create: true
    kind: SampleResource
`,
			expectedResult: `apiVersion: custom.example.org/v1
kind: SampleResource
metadata:
  labels:
    environment: dev
    location: planet-earth
  name: sample-resource
  namespace: sample-namespace
spec:
  selectors:
    labels:
      environment: dev
      location: planet-earth
  template:
    metadata:
      labels:
        environment: dev
        location: planet-earth
    spec:
      containers:
      - env:
        - name: VARIABLE
          value: value
        image: index.docker.io/library/hello-world
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			th := kusttest_test.MakeHarness(t)
			th.WriteK(".", tc.kustomization)
			th.WriteF("resources/custom-resource.yaml",
				`apiVersion: custom.example.org/v1
kind: SampleResource
metadata:
  name: sample-resource
  namespace: sample-namespace
spec:
  template:
    spec:
      containers:
      - image: index.docker.io/library/hello-world
        env:
        - name: VARIABLE
          value: value
`)

			th.WriteF("config/configurations.yaml", tc.transformerConfig)

			output := th.Run(".", th.MakeDefaultOptions())

			th.AssertActualEqualsExpected(output, tc.expectedResult)
		})
	}
}
