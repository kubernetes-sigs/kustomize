package main_test

/*
import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestHelmChartInflationGenerator(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("HelmChartInflationGenerator")
	defer th.Reset()

	rm := th.LoadAndRunGenerator(`
apiVersion: builtin
kind: HelmChartInflationGenerator
metadata:
  name: myMap
chartName: minecraft
chartRepoUrl: https://kubernetes-charts.storage.googleapis.com
chartVersion: v1.2.0
releaseName: test
releaseNamespace: testNamespace
`)

	th.AssertActualEqualsExpected(rm, `
apiVersion: v1
data:
  rcon-password: Q0hBTkdFTUUh
kind: Secret
metadata:
  labels:
    app: test-minecraft
    chart: minecraft-1.2.0
    heritage: Helm
    release: test
  name: test-minecraft
type: Opaque
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  annotations:
    volume.alpha.kubernetes.io/storage-class: default
  labels:
    app: test-minecraft
    chart: minecraft-1.2.0
    heritage: Helm
    release: test
  name: test-minecraft-datadir
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: test-minecraft
    chart: minecraft-1.2.0
    heritage: Helm
    release: test
  name: test-minecraft
spec:
  ports:
  - name: minecraft
    port: 25565
    protocol: TCP
    targetPort: minecraft
  selector:
    app: test-minecraft
  type: LoadBalancer
`)
}

func TestHelmChartInflationGeneratorWithValues(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		PrepBuiltin("HelmChartInflationGenerator")
	defer th.Reset()

	tempDirConfirmed, err := filesys.NewTmpConfirmedDir()
	if err != nil {
		t.Fatal(err)
	}
	tempDir := string(tempDirConfirmed)
	defer os.RemoveAll(tempDir)
	valuesPath := path.Join(tempDir, "values.yaml")
	ioutil.WriteFile(valuesPath, []byte(`
minecraftServer:
  eula: TRUE
`), 0644)

	rm := th.LoadAndRunGenerator(fmt.Sprintf(`
apiVersion: builtin
kind: HelmChartInflationGenerator
metadata:
  name: myMap
chartName: minecraft
chartRepoUrl: https://kubernetes-charts.storage.googleapis.com
chartVersion: v1.2.0
helmBin: helm
helmHome: %s
chartHome: %s
releaseName: test
releaseNamespace: testNamespace
values: %s
`, tempDir, tempDir, valuesPath))
valuesLocal:
  resources:
    limits:
      memory: 512Mi
      cpu: 1000m
    requests:
      memory: 512Mi
      cpu: 200m

	th.AssertActualEqualsExpected(rm, `
apiVersion: v1
data:
  rcon-password: Q0hBTkdFTUUh
kind: Secret
metadata:
  labels:
    app: test-minecraft
    chart: minecraft-1.2.0
    heritage: Helm
    release: test
  name: test-minecraft
type: Opaque
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  annotations:
    volume.alpha.kubernetes.io/storage-class: default
  labels:
    app: test-minecraft
    chart: minecraft-1.2.0
    heritage: Helm
    release: test
  name: test-minecraft-datadir
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app: test-minecraft
    chart: minecraft-1.2.0
    heritage: Helm
    release: test
  name: test-minecraft
spec:
  ports:
  - name: minecraft
    port: 25565
    protocol: TCP
    targetPort: minecraft
  selector:
    app: test-minecraft
  type: LoadBalancer
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: test-minecraft
    chart: minecraft-1.2.0
    heritage: Helm
    release: test
  name: test-minecraft
spec:
  selector:
    matchLabels:
      app: test-minecraft
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: test-minecraft
    spec:
      containers:
      - env:
        - name: EULA
          value: "true"
        - name: TYPE
          value: VANILLA
        - name: VERSION
          value: 1.14.4
        - name: DIFFICULTY
          value: easy
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
        - name: FORCE_gameMode
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
          value: 512M
        - name: JVM_OPTS
          value: ""
        - name: JVM_XX_OPTS
          value: ""
        image: itzg/minecraft-server:latest
        imagePullPolicy: Always
        livenessProbe:
          exec:
            command:
            - mcstatus
            - localhost:25565
            - status
          failureThreshold: 10
          initialDelaySeconds: 30
          periodSeconds: 5
          successThreshold: 1
          timeoutSeconds: 1
        name: test-minecraft
        ports:
        - containerPort: 25565
          name: minecraft
          protocol: TCP
        readinessProbe:
          exec:
            command:
            - mcstatus
            - localhost:25565
            - status
          failureThreshold: 10
          initialDelaySeconds: 30
          periodSeconds: 5
          successThreshold: 1
          timeoutSeconds: 1
        resources:
          requests:
            cpu: 500m
            memory: 512Mi
        volumeMounts:
        - mountPath: /data
          name: datadir
      securityContext:
        fsGroup: 2000
        runAsUser: 1000
      volumes:
      - name: datadir
        persistentVolumeClaim:
          claimName: test-minecraft-datadir
`)
}
*/
