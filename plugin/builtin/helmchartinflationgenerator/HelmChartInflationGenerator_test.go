package main_test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
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
	testData, err := ioutil.ReadFile("include_crds_testdata.txt")
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
