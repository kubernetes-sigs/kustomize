## Configuration Basics

### Synopsis

`kustomize config` provides tools for working with local configuration directories.

  First fetch a bundle of configuration to your local file system from the
  Kubernetes examples repository.

	git clone https://github.com/kubernetes/examples/
	cd examples/

### `tree` -- view Resources and directory structure

  `tree` can be used to summarize the collection of Resources in a directory:

	$ kustomize config tree mysql-wordpress-pd/
	mysql-wordpress-pd
	├── [gce-volumes.yaml]  v1.PersistentVolume wordpress-pv-1
	├── [gce-volumes.yaml]  v1.PersistentVolume wordpress-pv-2
	├── [local-volumes.yaml]  v1.PersistentVolume local-pv-1
	├── [local-volumes.yaml]  v1.PersistentVolume local-pv-2
	├── [mysql-deployment.yaml]  v1.PersistentVolumeClaim mysql-pv-claim
	├── [mysql-deployment.yaml]  apps/v1.Deployment wordpress-mysql
	├── [mysql-deployment.yaml]  v1.Service wordpress-mysql
	├── [wordpress-deployment.yaml]  apps/v1.Deployment wordpress
	├── [wordpress-deployment.yaml]  v1.Service wordpress
	└── [wordpress-deployment.yaml]  v1.PersistentVolumeClaim wp-pv-claim

  `tree` may be provided flags to print the Resource field values.  `tree` has a number of built-in
  supported fields, and may also print arbitrary values using the `--field` flag to specify a field
  path.

    $  kustomize config tree mysql-wordpress-pd/ --name --image --replicas --ports
    mysql-wordpress-pd
    ├── [gce-volumes.yaml]  PersistentVolume wordpress-pv-1
    ├── [gce-volumes.yaml]  PersistentVolume wordpress-pv-2
    ├── [local-volumes.yaml]  PersistentVolume local-pv-1
    ├── [local-volumes.yaml]  PersistentVolume local-pv-2
    ├── [mysql-deployment.yaml]  PersistentVolumeClaim mysql-pv-claim
    ├── [mysql-deployment.yaml]  Deployment wordpress-mysql
    │   └── spec.template.spec.containers
    │       └── 0
    │           ├── name: mysql
    │           ├── image: mysql:5.6
    │           └── ports: [{name: mysql, containerPort: 3306}]
    ├── [mysql-deployment.yaml]  Service wordpress-mysql
    │   └── spec.ports: [{port: 3306}]
    ├── [wordpress-deployment.yaml]  Deployment wordpress
    │   └── spec.template.spec.containers
    │       └── 0
    │           ├── name: wordpress
    │           ├── image: wordpress:4.8-apache
    │           └── ports: [{name: wordpress, containerPort: 80}]
    ├── [wordpress-deployment.yaml]  Service wordpress
    │   └── spec.ports: [{port: 80}]
    └── [wordpress-deployment.yaml]  PersistentVolumeClaim wp-pv-claim

  `tree` can also be used with `kubectl get` to print cluster Resources using OwnersReferences
  to build the tree structure.

    kubectl apply -R -f cockroachdb/
    kubectl get all -o yaml | kustomize config tree --graph-structure owners --name --image --replicas
    .
    ├── [Resource]  Deployment wp/wordpress
    │   ├── spec.replicas: 1
    │   ├── spec.template.spec.containers
    │   │   └── 0
    │   │       ├── name: wordpress
    │   │       └── image: wordpress:4.8-apache
    │   └── [Resource]  ReplicaSet wp/wordpress-76b5d9f5c8
    │       ├── spec.replicas: 1
    │       ├── spec.template.spec.containers
    │       │   └── 0
    │       │       ├── name: wordpress
    │       │       └── image: wordpress:4.8-apache
    │       └── [Resource]  Pod wp/wordpress-76b5d9f5c8-g656w
    │           └── spec.containers
    │               └── 0
    │                   ├── name: wordpress
    │                   └── image: wordpress:4.8-apache
    ├── [Resource]  Service wp/wordpress
    ...

### `cat` -- view the full collection of Resources

	$ kustomize config cat mysql-wordpress-pd/
	apiVersion: v1
	kind: PersistentVolume
	metadata:
	  name: wordpress-pv-1
	  annotations:
		config.kubernetes.io/path: gce-volumes.yaml
	spec:
	  accessModes:
	  - ReadWriteOnce
	  capacity:
		storage: 20Gi
	  gcePersistentDisk:
		fsType: ext4
		pdName: wordpress-1
	---
	apiVersion: v1
	...

  `cat` prints the raw package Resources.  This may be used to pipe them to other tools
  such as `kubectl apply -f -`.

## `fmt` -- format the Resources for a directory (like go fmt for Kubernetes Resources)

  `fmt` formats the Resource Configuration by applying a consistent style, including
  ordering of fields and indentation.

	$ kustomize config fmt mysql-wordpress-pd/

  Run `git diff` and see the changes that have been applied.

### `grep` -- search for Resources by field values

  `grep` prints Resources matching some field value.  The Resources are annotated with their
  file source so they can be piped to other commands without losing this information.

	$ kustomize config grep "metadata.name=wordpress" wordpress/
	apiVersion: v1
	kind: Service
	metadata:
	  name: wordpress
	  labels:
		app: wordpress
	  annotations:
		config.kubernetes.io/path: wordpress-deployment.yaml
	spec:
	  ports:
	  - port: 80
	  selector:
		app: wordpress
		tier: frontend
	  type: LoadBalancer
	---
	...

  - list elements may be indexed by a field value using list[field=value]
  - '.' as part of a key or value may be escaped as '\.'

	$ kustomize config grep "spec.status.spec.containers[name=nginx].image=mysql:5\.6" wordpress/
	apiVersion: apps/v1 # for k8s versions before 1.9.0 use apps/v1beta2  and before 1.8.0 use extensions/v1beta1
	kind: Deployment
	metadata:
	  name: wordpress-mysql
	  labels:
		app: wordpress
	spec:
	  selector:
		matchLabels:
		  app: wordpress
		  tier: mysql
	  template:
		metadata:
		  labels:
			app: wordpress
			tier: mysql
	...

  `grep` may be used with kubectl to search for Resources in a cluster matching a value.

    kubectl get all -o yaml | kustomize config grep "spec.replicas>0" | kustomize config tree --replicas
    .
    └──
        ├── [.]  Deployment wp/wordpress
        │   └── spec.replicas: 1
        ├── [.]  ReplicaSet wp/wordpress-76b5d9f5c8
        │   └── spec.replicas: 1
        ├── [.]  Deployment wp/wordpress-mysql
        │   └── spec.replicas: 1
        └── [.]  ReplicaSet wp/wordpress-mysql-f9447f458
            └── spec.replicas: 1

### Error handling

  If there is an error parsing the Resource configuration, kustomize will print an error with the file.

    $ kustomize config grep "spec.template.spec.containers[name=\.*].resources.limits.cpu>1.0" ./staging/ | kustomize config tree --name --resources
    Error: staging/persistent-volume-provisioning/quobyte/quobyte-admin-secret.yaml: [0]: yaml: unmarshal errors:
      line 13: mapping key "type" already defined at line 9

  Here the `staging/persistent-volume-provisioning/quobyte/quobyte-admin-secret.yaml` has a malformed
  Resource.  Remove the malformed Resources:

    rm staging/persistent-volume-provisioning/quobyte/quobyte-admin-secret.yaml
    rm staging/storage/vitess/etcd-service-template.yaml

  When developing -- to get a stack trace for where an error was encountered,
  use the `--stack-trace` flag:

    $ kustomize config grep "spec.template.spec.containers[name=\.*].resources.limits.cpu>1.0" ./staging/ --stack-trace
    go/src/sigs.k8s.io/kustomize/kyaml/yaml/types.go:260 (0x4d35c86)
            (*RNode).GetMeta: return m, errors.Wrap(err)
    go/src/sigs.k8s.io/kustomize/kyaml/kio/byteio_reader.go:130 (0x4d3e099)
            (*ByteReader).Read: meta, err := node.GetMeta()
    ...


### Combine `grep` and `tree`

  `grep` and `tree` may be combined to perform queries against configuration.

  Query for `replicas`:

    $ kustomize config grep "spec.replicas>5" ./ | kustomize config tree --replicas
      .
      ├── staging/sysdig-cloud
      │   └── [sysdig-rc.yaml]  ReplicationController sysdig-agent
      │       └── spec.replicas: 100
      └── staging/volumes/vsphere
          └── [simple-statefulset.yaml]  StatefulSet web
              └── spec.replicas: 14

  Query for `resource.limits`

	$ kustomize config grep "spec.template.spec.containers[name=\.*].resources.limits.memory>0" ./ | kustomize config tree --resources
	.
    ├── cassandra
    │   └── [cassandra-statefulset.yaml]  StatefulSet cassandra
    │       └── spec.template.spec.containers
    │           └── 0
    │               └── resources: {limits: {cpu: "500m", memory: 1Gi}, requests: {cpu: "500m", memory: 1Gi}}
    ├── staging/selenium
    │   ├── [selenium-hub-deployment.yaml]  Deployment selenium-hub
    │   │   └── spec.template.spec.containers
    │   │       └── 0
    │   │           └── resources: {limits: {memory: 1000Mi, cpu: ".5"}}
    │   ├── [selenium-node-chrome-deployment.yaml]  Deployment selenium-node-chrome
    │   │   └── spec.template.spec.containers
    │   │       └── 0
    │   │           └── resources: {limits: {memory: 1000Mi, cpu: ".5"}}
    │   └── [selenium-node-firefox-deployment.yaml]  Deployment selenium-node-firefox
    │       └── spec.template.spec.containers
    │           └── 0
    │               └── resources: {limits: {memory: 1000Mi, cpu: ".5"}}
    ...

### Inverting `grep`

  The `grep` results may be inverted with the `-v` flag and used to find Resources that don't
  match a query.

  Find Resources that have an image specified, but the image doesn't have a tag:

    $ kustomize config grep "spec.template.spec.containers[name=\.*].name=\.*" ./ |  kustomize config grep "spec.template.spec.containers[name=\.*].image=\.*:\.*" -v | kustomize config tree --image --name
    .
    ├── staging/newrelic
    │   ├── [newrelic-daemonset.yaml]  DaemonSet newrelic-agent
    │   │   └── spec.template.spec.containers
    │   │       └── 0
    │   │           ├── name: newrelic
    │   │           └── image: newrelic/nrsysmond
    │   └── staging/newrelic-infrastructure
    │       └── [newrelic-infra-daemonset.yaml]  DaemonSet newrelic-infra-agent
    │           └── spec.template.spec.containers
    │               └── 0
    │                   ├── name: newrelic
    │                   └── image: newrelic/infrastructure
    ├── staging/nodesjs-mongodb
    │   ├── [mongo-controller.yaml]  ReplicationController mongo-controller
    │   │   └── spec.template.spec.containers
    │   │       └── 0
    │   │           ├── name: mongo
    │   │           └── image: mongo
    │   └── [web-controller.yaml]  ReplicationController web-controller
    │       └── spec.template.spec.containers
    │           └── 0
    │               ├── name: web
    │               └── image: <YOUR-CONTAINER>
    ...
