package main_test

import (
	"fmt"
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
  name: myMc
name: minecraft
version: 3.1.3
repo: https://itzg.github.io/minecraft-server-charts
releaseName: moria
`)

	th.AssertActualEqualsExpected(rm, `
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
