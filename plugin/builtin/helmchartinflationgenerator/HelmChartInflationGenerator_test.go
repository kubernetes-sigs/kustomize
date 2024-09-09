// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
	"sigs.k8s.io/kustomize/kyaml/copyutil"
)

func TestHelmChartInflationGenerator(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarnessWithTmpRoot(t).
		PrepBuiltin("HelmChartInflationGenerator")
	defer th.Reset()
	if err := th.ErrIfNoHelm(); err != nil {
		t.Skip("skipping: " + err.Error())
	}

	rm := th.LoadAndRunGenerator(`
apiVersion: builtin
kind: HelmChartInflationGenerator
metadata:
  name: myPipeline
name: ocp-pipeline
namespace: mynamespace
version: 0.1.16
repo: https://bcgov.github.io/helm-charts
releaseName: moria
valuesInline:
  releaseNamespace: ""
  rbac:
    create: true
    rules:
      - apiGroups: [""]
        verbs: ["*"]
        resouces: ["*"]
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: v1
data:
  config: eyJleGFtcGxlIjoidmFsdWUifQ==
kind: Secret
metadata:
  labels:
    chart: ocp-pipeline-0.1.16
    heritage: Helm
    release: moria
  name: moria-config
type: Opaque
---
apiVersion: v1
data:
  WebHookSecretKey: MTIzNDU2Nzg=
kind: Secret
metadata:
  labels:
    chart: ocp-pipeline-0.1.16
    heritage: Helm
    release: moria
  name: moria-git-webhook-secret
type: Opaque
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: moria-ocp-pipeline
  namespace: mynamespace
rules:
- apiGroups:
  - ""
  resouces:
  - '*'
  verbs:
  - '*'
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: moria-ocp-pipeline
  namespace: mynamespace
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: moria-ocp-pipeline
subjects:
- kind: ServiceAccount
  name: jenkins
  namespace: mynamespace
---
apiVersion: build.openshift.io/v1
kind: BuildConfig
metadata:
  labels:
    app: ocp-pipeline
    chart: ocp-pipeline-0.1.16
    heritage: Helm
    release: moria
  name: moria-ocp-pipeline-deploy
  namespace: null
spec:
  nodeSelector: {}
  resources:
    limits:
      cpu: 4000m
      memory: 8G
    requests:
      cpu: 2000m
      memory: 4G
  strategy:
    jenkinsPipelineStrategy:
      jenkinsfile: |-
        def helmName = "helm-v3.1.0-linux-amd64.tar.gz"
        def chartName = "metadata-curator"
        def chartRepo = "http://bcgov.github.io/helm-charts"
        def releaseName  = "mc"
        def releaseNamespace = ""
        def forceRecreate = "false"
        def callAnotherPipe = "false"
        def useEnv = "false"
        def fromEnv = "commit"
        def setFlag = "image.tag"

          node("nodejs") {
            stage("deploy (it's already built)") {
              sh """
                curl -L -O https://get.helm.sh/${helmName}
                tar -zxvf ${helmName}
                cd linux-amd64

                curl -L -O https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux32
                chmod ugo+x ./jq-linux32
                npm install -g json2yaml

                export CONF1=`+"`"+`oc get secret moria-config -o json | ./jq-linux32 .data.config`+"`"+`
                export CONF2=`+"`"+`sed -e 's/^"//' -e 's/"\$//' <<<"\$CONF1"`+"`"+`
                export CONF3=`+"`"+`echo \$CONF2 | base64 -d -`+"`"+`
                export CONF=`+"`"+`echo \$CONF3 | json2yaml`+"`"+`

                echo "\$CONF" > ./config.yaml
                oc project ${releaseNamespace}
                ./helm repo add chart ${chartRepo}
                ./helm repo update
                if [ "${forceRecreate}" = "true" ]; then
                  ./helm upgrade ${releaseName} chart/${chartName} -f ./config.yaml --install --set hashLabel="${releaseName}\$(cat /dev/urandom | tr -dc 'a-z0-9' | fold -w 32 | head -n 1)"
                elif [ "${useEnv}" = "true" ]; then
                  ./helm upgrade ${releaseName} chart/${chartName} -f ./config.yaml --install --set ${setFlag}=${env[fromEnv]}
                else
                  ./helm upgrade ${releaseName} chart/${chartName} -f ./config.yaml --install
                fi

                if [ "${callAnotherPipe}" = "true" ]; then
                  curl -d '' http://otherwebhookUrl
                fi
              """
          }
        }
    type: JenkinsPipeline
  triggers:
  - generic:
      allowEnv: true
      secretReference:
        name: moria-git-webhook-secret
    type: generic
status:
  lastVersion: 0
`)
}

const expectedInflationFmt = `
apiVersion: v1
data:
  rcon-password: Q0hBTkdFTUUh
kind: Secret
metadata:
  labels:
    app: moria-minecraft
    chart: minecraft-3.1.3
    heritage: Helm
    release: moria
  name: moria-minecraft
type: Opaque
---
apiVersion: v1
kind: Service
metadata:
  annotations: {}
  labels:
    app: moria-minecraft
    chart: minecraft-3.1.3
    heritage: Helm
    release: moria
  name: moria-minecraft
spec:
  ports:
  - name: minecraft
    port: 25565
    protocol: TCP
    targetPort: minecraft
  selector:
    app: moria-minecraft
  type: ClusterIP
---
apiVersion: v1
kind: Service
metadata:
  annotations: {}
  labels:
    app: moria-minecraft
    chart: minecraft-3.1.3
    heritage: Helm
    release: moria
  name: moria-minecraft-rcon
spec:
  ports:
  - name: rcon
    port: 25575
    protocol: TCP
    targetPort: rcon
  selector:
    app: moria-minecraft
  type: LoadBalancer
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: moria-minecraft
    chart: minecraft-3.1.3
    heritage: Helm
    release: moria
  name: moria-minecraft
spec:
  selector:
    matchLabels:
      app: moria-minecraft
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: moria-minecraft
    spec:
      containers:
      - env:
        - name: EULA
          value: "true"
        - name: TYPE
          value: VANILLA
        - name: VERSION
          value: LATEST
        - name: DIFFICULTY
          value: %s
        - name: WHITELIST
          value: ""
        - name: OPS
          value: ""
        - name: ICON
          value: ""
        - name: MAX_PLAYERS
          value: "20"
        - name: MAX_WORLD_SIZE
          value: "10000"
        - name: ALLOW_NETHER
          value: "true"
        - name: ANNOUNCE_PLAYER_ACHIEVEMENTS
          value: "true"
        - name: ENABLE_COMMAND_BLOCK
          value: "true"
        - name: FORCE_GAMEMODE
          value: "false"
        - name: GENERATE_STRUCTURES
          value: "true"
        - name: HARDCORE
          value: "false"
        - name: MAX_BUILD_HEIGHT
          value: "256"
        - name: MAX_TICK_TIME
          value: "60000"
        - name: SPAWN_ANIMALS
          value: "true"
        - name: SPAWN_MONSTERS
          value: "true"
        - name: SPAWN_NPCS
          value: "true"
        - name: VIEW_DISTANCE
          value: "10"
        - name: SEED
          value: ""
        - name: MODE
          value: survival
        - name: MOTD
          value: Welcome to Minecraft on Kubernetes!
        - name: PVP
          value: "false"
        - name: LEVEL_TYPE
          value: DEFAULT
        - name: GENERATOR_SETTINGS
          value: ""
        - name: LEVEL
          value: world
        - name: ONLINE_MODE
          value: "true"
        - name: MEMORY
          value: 1024M
        - name: JVM_OPTS
          value: ""
        - name: JVM_XX_OPTS
          value: ""
        - name: ENABLE_RCON
          value: "true"
        - name: RCON_PASSWORD
          valueFrom:
            secretKeyRef:
              key: rcon-password
              name: moria-minecraft
        image: itzg/minecraft-server:latest
        imagePullPolicy: Always
        livenessProbe:
          failureThreshold: 10
          initialDelaySeconds: 30
          periodSeconds: 5
          successThreshold: 1
          tcpSocket:
            port: 25565
          timeoutSeconds: 1
        name: moria-minecraft
        ports:
        - containerPort: 25565
          name: minecraft
          protocol: TCP
        - containerPort: 25575
          name: rcon
          protocol: TCP
        readinessProbe:
          failureThreshold: 10
          initialDelaySeconds: 30
          periodSeconds: 5
          successThreshold: 1
          tcpSocket:
            port: 25565
          timeoutSeconds: 1
        resources:
          requests:
            cpu: %dm
            memory: %dMi
        volumeMounts:
        - mountPath: /data
          name: datadir
      securityContext:
        fsGroup: 2000
        runAsUser: 1000
      volumes:
      - emptyDir: {}
        name: datadir
`

func TestHelmChartInflationGeneratorWithValuesInlineOverride(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarnessWithTmpRoot(t).
		PrepBuiltin("HelmChartInflationGenerator")
	defer th.Reset()
	if err := th.ErrIfNoHelm(); err != nil {
		t.Skip("skipping: " + err.Error())
	}
	rm := th.LoadAndRunGenerator(`
apiVersion: builtin
kind: HelmChartInflationGenerator
metadata:
  name: myMc
name: minecraft
version: 3.1.3
repo: https://itzg.github.io/minecraft-server-charts
releaseName: moria
valuesInline:
  minecraftServer:
    eula: true
    difficulty: hard
    rcon:
      enabled: true
`)
	th.AssertActualEqualsExpected(
		rm, fmt.Sprintf(expectedInflationFmt,
			"hard", // difficulty
			500,    // cpu
			512,    // memory
		))
}

func TestHelmChartInflationGeneratorWithLocalValuesFile(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarnessWithTmpRoot(t).
		PrepBuiltin("HelmChartInflationGenerator")
	defer th.Reset()
	if err := th.ErrIfNoHelm(); err != nil {
		t.Skip("skipping: " + err.Error())
	}
	th.WriteF(filepath.Join(th.GetRoot(), "myValues.yaml"), `
minecraftServer:
  eula: true
  difficulty: peaceful
  rcon:
    enabled: true
resources:
  requests:
    cpu: 888m
    memory: 666Mi
`)
	rm := th.LoadAndRunGenerator(`
apiVersion: builtin
kind: HelmChartInflationGenerator
metadata:
  name: myMc
name: minecraft
version: 3.1.3
repo: https://itzg.github.io/minecraft-server-charts
releaseName: moria
valuesFile: myValues.yaml
`)
	th.AssertActualEqualsExpected(
		rm, fmt.Sprintf(expectedInflationFmt,
			"peaceful", // difficulty
			888,        // cpu
			666,        // memory
		))
}

func TestHelmChartInflationGeneratorWithInLineReplace(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarnessWithTmpRoot(t).
		PrepBuiltin("HelmChartInflationGenerator")
	defer th.Reset()
	if err := th.ErrIfNoHelm(); err != nil {
		t.Skip("skipping: " + err.Error())
	}
	rm := th.LoadAndRunGenerator(`
apiVersion: builtin
kind: HelmChartInflationGenerator
metadata:
  name: myMc
name: minecraft
version: 3.1.3
repo: https://itzg.github.io/minecraft-server-charts
releaseName: moria
valuesInline:
  minecraftServer:
    eula: true
    difficulty: OMG_PLEASE_NO
    rcon:
      enabled: true
  resources:
    requests:
      cpu: 555m
      memory: 111Mi
valuesMerge: replace
`)
	th.AssertActualEqualsExpected(
		rm, fmt.Sprintf(expectedInflationFmt,
			"OMG_PLEASE_NO", // difficulty
			555,             // cpu
			111,             // memory
		))
}

func TestHelmChartInflationGeneratorWithIncludeCRDs(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarnessWithTmpRoot(t).
		PrepBuiltin("HelmChartInflationGenerator")
	defer th.Reset()
	if err := th.ErrIfNoHelm(); err != nil {
		t.Skip("skipping: " + err.Error())
	}

	// we store this data outside of the _test.go file as its sort of huge
	// and has backticks, which makes string literals wonky
	testData, err := os.ReadFile("include_crds_testdata.txt")
	if err != nil {
		t.Error(fmt.Errorf("unable to read test data for includeCRDs: %w", err))
	}

	rm := th.LoadAndRunGenerator(`
apiVersion: builtin
kind: HelmChartInflationGenerator
metadata:
  name: terraform
name: terraform
version: 1.0.0
repo: https://helm.releases.hashicorp.com
releaseName: terraforming-mars
includeCRDs: true
valuesInline:
  global:
    enabled: false
  tests:
    enabled: false
`)
	th.AssertActualEqualsExpected(rm, string(testData))
}

func TestHelmChartInflationGeneratorWithExcludeCRDs(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarnessWithTmpRoot(t).
		PrepBuiltin("HelmChartInflationGenerator")
	defer th.Reset()
	if err := th.ErrIfNoHelm(); err != nil {
		t.Skip("skipping: " + err.Error())
	}

	// we choose this helm chart as it has the ability to turn
	// everything off, except CRDs.
	rm := th.LoadAndRunGenerator(`
apiVersion: builtin
kind: HelmChartInflationGenerator
metadata:
  name: terraform
name: terraform
version: 1.0.0
repo: https://helm.releases.hashicorp.com
releaseName: terraforming-mars
includeCRDs: false
valuesInline:
  global:
    enabled: false
  tests:
    enabled: false
`)
	th.AssertActualEqualsExpected(rm, "")
}

func TestHelmChartInflationGeneratorWithSkipHooks(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarnessWithTmpRoot(t).
		PrepBuiltin("HelmChartInflationGenerator")
	defer th.Reset()
	if err := th.ErrIfNoHelm(); err != nil {
		t.Skip("skipping: " + err.Error())
	}

	// we choose this helm chart as it has the ability to turn
	// everything off, except CRDs.
	rm := th.LoadAndRunGenerator(`
apiVersion: builtin
kind: HelmChartInflationGenerator
metadata:
  name: terraform
name: terraform
version: 1.0.0
repo: https://helm.releases.hashicorp.com
releaseName: terraforming-mars
includeCRDs: false
skipHooks: true
valuesInline:
  global:
    enabled: false
`)
	th.AssertActualEqualsExpected(rm, "")
}

func TestHelmChartInflationGeneratorWithIncludeCRDsNotSpecified(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarnessWithTmpRoot(t).
		PrepBuiltin("HelmChartInflationGenerator")
	defer th.Reset()
	if err := th.ErrIfNoHelm(); err != nil {
		t.Skip("skipping: " + err.Error())
	}

	rm := th.LoadAndRunGenerator(`
apiVersion: builtin
kind: HelmChartInflationGenerator
metadata:
  name: terraform
name: terraform
version: 1.0.0
repo: https://helm.releases.hashicorp.com
releaseName: terraforming-mars
valuesInline:
  global:
    enabled: false
  tests:
    enabled: false
`)
	th.AssertActualEqualsExpected(rm, "")
}

func TestHelmChartInflationGeneratorIssue4905(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarnessWithTmpRoot(t).
		PrepBuiltin("HelmChartInflationGenerator")
	defer th.Reset()
	if err := th.ErrIfNoHelm(); err != nil {
		t.Skip("skipping: " + err.Error())
	}

	copyTestChartsIntoHarness(t, th)

	rm := th.LoadAndRunGenerator(`
apiVersion: builtin
kind: HelmChartInflationGenerator
metadata:
  name: issue4905
name: issue4905
releaseName: issue4905
chartHome: ./charts
valuesInline:
  config:
    item1: 1
    item2: 2
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: v1
data:
  config.yaml: |-
    item1: 1
    item2: 2
kind: ConfigMap
metadata:
  name: issue4905
`)
}

func TestHelmChartInflationGeneratorValuesOverride(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarnessWithTmpRoot(t).
		PrepBuiltin("HelmChartInflationGenerator")
	defer th.Reset()
	if err := th.ErrIfNoHelm(); err != nil {
		t.Skip("skipping: " + err.Error())
	}

	copyTestChartsIntoHarness(t, th)

	rm := th.LoadAndRunGenerator(`
apiVersion: builtin
kind: HelmChartInflationGenerator
metadata:
  name: values-merge
name: values-merge
releaseName: values-merge
valuesMerge: override
valuesInline:
  a: 4
  c: 3
  list:
  - c
  map:
    a: 7
    c: 6
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: test.kustomize.io/v1
kind: ValuesMergeTest
metadata:
  name: values-merge
obj:
  a: 4
  b: 2
  c: 3
  list:
  - c
  map:
    a: 7
    b: 5
    c: 6
`)
}

func TestHelmChartInflationGeneratorValuesReplace(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarnessWithTmpRoot(t).
		PrepBuiltin("HelmChartInflationGenerator")
	defer th.Reset()
	if err := th.ErrIfNoHelm(); err != nil {
		t.Skip("skipping: " + err.Error())
	}

	copyTestChartsIntoHarness(t, th)

	th.WriteF(filepath.Join(th.GetRoot(), "replacedValues.yaml"), `
  a: 7
  b: 7
  c: 7
  list:
  - g
  map:
    a: 7
    b: 7
    c: 7
  `)

	rm := th.LoadAndRunGenerator(`
apiVersion: builtin
kind: HelmChartInflationGenerator
metadata:
  name: values-merge
name: values-merge
releaseName: values-merge
valuesMerge: replace
valuesFile: replacedValues.yaml
valuesInline:
  a: 1
  b: 2
  list:
  - c
  map:
    a: 4
    b: 5
`)

	// replace option does not ignore values file from the chart
	// it only replaces the values files specified in the kustomization
	th.AssertActualEqualsExpected(rm, `
apiVersion: test.kustomize.io/v1
kind: ValuesMergeTest
metadata:
  name: values-merge
obj:
  a: 1
  b: 2
  c: null
  list:
  - c
  map:
    a: 4
    b: 5
    c: null
`)
}

func TestHelmChartInflationGeneratorValuesMerge(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarnessWithTmpRoot(t).
		PrepBuiltin("HelmChartInflationGenerator")
	defer th.Reset()
	if err := th.ErrIfNoHelm(); err != nil {
		t.Skip("skipping: " + err.Error())
	}

	copyTestChartsIntoHarness(t, th)

	rm := th.LoadAndRunGenerator(`
apiVersion: builtin
kind: HelmChartInflationGenerator
metadata:
  name: values-merge
name: values-merge
releaseName: values-merge
valuesMerge: merge
valuesInline:
  a: 4
  c: 3
  list:
  - c
  map:
    a: 7
    c: 6
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: test.kustomize.io/v1
kind: ValuesMergeTest
metadata:
  name: values-merge
obj:
  a: 1
  b: 2
  c: 3
  list:
  - a
  - b
  map:
    a: 4
    b: 5
    c: 6
`)
}

func copyTestChartsIntoHarness(t *testing.T, th *kusttest_test.HarnessEnhanced) {
	t.Helper()

	thDir := filepath.Join(th.GetRoot(), "charts")
	chartDir := "testdata/charts"

	require.NoError(t, copyutil.CopyDir(th.GetFSys(), chartDir, thDir))
}

func TestHelmChartInflationGeneratorWithSameChartMultipleVersions(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarnessWithTmpRoot(t).
		PrepBuiltin("HelmChartInflationGenerator")
	defer th.Reset()
	if err := th.ErrIfNoHelm(); err != nil {
		t.Skip("skipping: " + err.Error())
	}
	tests := []struct {
		name        string
		chartName   string
		repo        string
		version     string
		releaseName string
	}{
		{
			name:        "terraform chart with no version grabs latest",
			chartName:   "terraform",
			repo:        "https://helm.releases.hashicorp.com",
			version:     "",
			releaseName: "terraform-latest",
		},
		{
			name:        "terraform chart with version 1.1.1",
			chartName:   "terraform",
			repo:        "https://helm.releases.hashicorp.com",
			version:     "1.1.1",
			releaseName: "terraform-1.1.1",
		},
		{
			name:        "terraform chart with version 1.1.1 again",
			chartName:   "terraform",
			repo:        "https://helm.releases.hashicorp.com",
			version:     "1.1.1",
			releaseName: "terraform-1.1.1-1",
		},
		{
			name:        "terraform chart with version 1.1.2",
			chartName:   "terraform",
			repo:        "https://helm.releases.hashicorp.com",
			version:     "1.1.2",
			releaseName: "terraform-1.1.2",
		},
		{
			name:        "terraform chart with version 1.1.2 but without repo",
			chartName:   "terraform",
			version:     "1.1.2",
			releaseName: "terraform-1.1.2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := fmt.Sprintf(`
apiVersion: builtin
kind: HelmChartInflationGenerator
metadata:
 name: %s
name: %s
version: %s
repo: %s
releaseName: %s
`, tt.chartName, tt.chartName, tt.version, tt.repo, tt.releaseName)

			rm := th.LoadAndRunGenerator(config)
			assert.True(t, len(rm.Resources()) > 0)

			var chartDir string
			if tt.version != "" && tt.repo != "" {
				chartDir = fmt.Sprintf("charts/%s-%s/%s", tt.chartName, tt.version, tt.chartName)
			} else {
				chartDir = fmt.Sprintf("charts/%s", tt.chartName)
			}

			fmt.Printf("%s: %s\n", tt.name, chartDir)

			d, err := th.GetFSys().ReadFile(filepath.Join(th.GetRoot(), chartDir, "Chart.yaml"))
			if err != nil {
				t.Fatal(err)
			}

			assert.Contains(t, string(d), fmt.Sprintf("name: %s", tt.chartName))
			if tt.version != "" {
				assert.Contains(t, string(d), fmt.Sprintf("version: %s", tt.version))
			}
		})
	}
}

// Test that verifies +1 instances of same chart with different versions
// https://github.com/kubernetes-sigs/kustomize/issues/4813
func TestHelmChartInflationGeneratorWithMultipleInstancesSameChartDifferentVersions(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarnessWithTmpRoot(t).
		PrepBuiltin("HelmChartInflationGenerator")
	defer th.Reset()
	if err := th.ErrIfNoHelm(); err != nil {
		t.Skip("skipping: " + err.Error())
	}

	podinfo1 := th.LoadAndRunGenerator(`
apiVersion: builtin
kind: HelmChartInflationGenerator
metadata:
 name: podinfo
name: podinfo
version: 6.2.1
repo: https://stefanprodan.github.io/podinfo
releaseName: podinfo1
`)

	podinfo2 := th.LoadAndRunGenerator(`
apiVersion: builtin
kind: HelmChartInflationGenerator
metadata:
 name: podinfo
name: podinfo
version: 6.1.8
repo: https://stefanprodan.github.io/podinfo
releaseName: podinfo2
`)

	podinfo1Img, err := podinfo1.Resources()[1].GetFieldValue("spec.template.spec.containers.0.image")
	require.NoError(t, err)
	assert.Equal(t, "ghcr.io/stefanprodan/podinfo:6.2.1", podinfo1Img)

	podinfo2Img, err := podinfo2.Resources()[1].GetFieldValue("spec.template.spec.containers.0.image")
	require.NoError(t, err)
	assert.Equal(t, "ghcr.io/stefanprodan/podinfo:6.1.8", podinfo2Img)

	podinfo1ChartsDir := filepath.Join(th.GetRoot(), "charts/podinfo-6.2.1/podinfo")
	assert.True(t, th.GetFSys().Exists(podinfo1ChartsDir))

	podinfo2ChartsDir := filepath.Join(th.GetRoot(), "charts/podinfo-6.1.8/podinfo")
	assert.True(t, th.GetFSys().Exists(podinfo2ChartsDir))

	podinfo1ChartContents, err := th.GetFSys().ReadFile(filepath.Join(podinfo1ChartsDir, "Chart.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(podinfo1ChartContents), "version: 6.2.1")

	podinfo2ChartContents, err := th.GetFSys().ReadFile(filepath.Join(podinfo2ChartsDir, "Chart.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(podinfo2ChartContents), "version: 6.1.8")
}

// Reference: https://github.com/kubernetes-sigs/kustomize/issues/5163
func TestHelmChartInflationGeneratorUsingVersionWithoutRepo(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarnessWithTmpRoot(t).
		PrepBuiltin("HelmChartInflationGenerator")
	defer th.Reset()
	if err := th.ErrIfNoHelm(); err != nil {
		t.Skip("skipping: " + err.Error())
	}

	copyTestChartsIntoHarness(t, th)

	rm := th.LoadAndRunGenerator(`
apiVersion: builtin
kind: HelmChartInflationGenerator
metadata:
  name: test-chart
name: test-chart
version: 1.0.0
releaseName: test
chartHome: ./charts
`)

	cm, err := rm.Resources()[0].GetFieldValue("metadata.name")
	require.NoError(t, err)
	assert.Equal(t, "bar", cm)

	chartDir := filepath.Join(th.GetRoot(), "charts/test-chart")
	assert.True(t, th.GetFSys().Exists(chartDir))

	chartYamlContent, err := th.GetFSys().ReadFile(filepath.Join(chartDir, "Chart.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(chartYamlContent), "name: test-chart")
	assert.Contains(t, string(chartYamlContent), "version: 1.0.0")
}

func TestHelmChartInflationGeneratorWithDebug(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarnessWithTmpRoot(t).
		PrepBuiltin("HelmChartInflationGenerator")
	defer th.Reset()
	if err := th.ErrIfNoHelm(); err != nil {
		t.Skip("skipping: " + err.Error())
	}
	copyTestChartsIntoHarness(t, th)

	rm := th.LoadAndRunGenerator(`
apiVersion: builtin
kind: HelmChartInflationGenerator
metadata:
  name: test-chart
name: test-chart
version: 1.0.0
releaseName: test
chartHome: ./charts
debug: true
`)

	cm, err := rm.Resources()[0].GetFieldValue("metadata.name")
	require.NoError(t, err)
	assert.Equal(t, "bar", cm)

	chartDir := filepath.Join(th.GetRoot(), "charts/test-chart")
	assert.True(t, th.GetFSys().Exists(chartDir))

	chartYamlContent, err := th.GetFSys().ReadFile(filepath.Join(chartDir, "Chart.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(chartYamlContent), "name: test-chart")
	assert.Contains(t, string(chartYamlContent), "version: 1.0.0")
}
