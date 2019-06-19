# inventory directive in kustomization.yaml

New in v2.1.0, a kustomization file may have an `inventory` field:
```yaml
inventory:
  type: ConfigMap
  configMap:
    name: prune-cm-name
    namespace: some-namespace
```

### Motivation

If present, `kustomize build` will make an _inventory_ object,
which  could be a ConfigMap, or an App (to be added),
which can be consumed by a client such as those under development in
[cli-experimental](https://github.com/kubernetes-sigs/cli-experimental).

The client can recognize this object by name and use it to do a better job
with actions like `apply`, `prune` and `delete`.


### Implementation

The _inventory_ ConfigMap contains two special annotations:

- kustomize.config.k8s.io/Inventory
  The value of this annotation is the JSON blob
  for an Inventory object. The Inventory is a
  struct that contains following information
  - all objects within this kustomization target
  - all objects that reference within this kustomization target

  Here is an example of an Inventory object
  ```json
  {
    "current":
      {
        "apps_v1_Deployment|default|mysql":null,
        "~G_v1_Secret|default|pass-dfg7h97cf6":
          [
            {
              "group":"apps",
              "version":"v1",
              "kind":"Deployment",
              "name":"mysql",
              "namespace":"default"
            }
          ],
        "~G_v1_Service|default|mysql":null
      }
    }
  ```

- kustomize.config.k8s.io/InventoryHash
  The value of this annotation is a hash that is
  computed from the list of items in the Inventory

Basically, this inventory object acts a record of objects that are applied as a group.
This object can be consumed by a client such as
[cli-experimental](https://github.com/kubernetes-sigs/cli-experimental).
The client can recognize the inventory annotations and take proper actions
when running apply, prune and delete.

### Example
Take following `kustomization.yaml` as an example
```yaml
resources:
- deployment.yaml
- service.yaml


secretGenerator:
- name: pass
  literals:
  - password=secret

inventory:
  type: ConfigMap
  configMap:
    name: root-cm
    namespace: default

namespace: default
```

where the `deployment.yaml` is
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mysql
  labels:
    app: mysql
spec:
  revisionHistoryLimit: 2
  selector:
    matchLabels:
      app: mysql
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: mysql
    spec:
      containers:
      - image: mysql:5.6
        name: mysql
        env:
        - name: MYSQL_ROOT_PASSWORD
          valueFrom:
            secretKeyRef:
              name: pass
              key: password
        ports:
        - containerPort: 3306
          name: mysql
        volumeMounts:
        - name: mysql-persistent-storage
          mountPath: /var/lib/mysql
      volumes:
      - name: mysql-persistent-storage
        emptyDir: {}
```

and the `service.yaml` is
```yaml
apiVersion: v1
kind: Service
metadata:
  name: mysql
  labels:
    app: mysql
spec:
  ports:
    - port: 3306
  selector:
    app: mysql
```

Running `kustomize build` gives 4 objects.
Besides the Deployment `mysql`, the Service `mysql`,
and the Secret `pass`, the output also contains a
ConfigMap object as
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  annotations:
    kustomize.config.k8s.io/Inventory: '{"current":{"apps_v1_Deployment|default|mysql":null,"~G_v1_Secret|default|pass-dfg7h97cf6":[{"group":"apps","version":"v1","kind":"Deployment","name":"mysql","namespace":"default"}],"~G_v1_Service|default|mysql":null}}'
    kustomize.config.k8s.io/InventoryHash: 7mgt867b75
  name: root-cm
  namespace: default
```

It is clear that this ConfigMap contains an `Inventory` annotation.


### Hash
Note that in the ConfigMap generated from `inventory` field, there is a hash
`b965tb9c7d`. It is the value for annotation `kustomize.config.k8s.io/InventoryHash`.

This hash is computed by hashing all the keys in data field, which is the following list
in this example.
```yaml
apps_v1_Deployment|default|mysql
~G_v1_Secret|default|pass-dfg7h97cf6
~G_v1_Service|default|mysql
```
When any object is added or removed from the kustomzation target, the hash changes. Thus by simply comparing the hash in the inventory objects, one can determine if the list of objects has changed.


### How prune works
In [cli-experimental](https://github.com/kubernetes-sigs/cli-experimental), there are different subcommands, `apply` and `prune`. Both are able to recognize an _inventory_ object and looking for its existing object on the cluster.

the `apply` command
recognizes the _inventory_ object by the annotation `kustomize.config.k8s.io/InventoryHash`. It then compares the current hash with the hash for the same object in the cluster. Since the hash reflects if there is any object added or removed, `apply` takes different actions correspondingly.
- When there is no existing _inventory_ object in the cluster, apply creates the inventory object.
- When the current hash is the same as the one in cluster, apply doesn't change the existing object in the cluster.
- when the current hash is different, apply merges the inventory annotation of the existing object in the cluster and the incoming object. The hash is updated to the latest hash.


The `prune` command parses the value of `kustomize.config.k8s.io/Inventory` of the existing _inventory_ object and computes two sets of objects based on the parsed data.
To be simple,
- The items in `Inventory.Current` will be kept
- The items in `Inventory.Previous` will be pruned when they
  are not needed.

