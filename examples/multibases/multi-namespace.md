# Demo: multi namespaces with a common base

`kustomize` supports defining multiple variants with different namespace, as overlays on a common base.

It's possible to create an additional overlay to compose these variants
together - just declare the overlays as the bases of a new kustomization. The
following demonstrates this using a base that's just one pod.

Define a place to work:

<!-- @makeWorkplace @testAgainstLatestRelease -->
```
DEMO_HOME=$(mktemp -d)
```

Define a common base:
<!-- @makeBase @testAgainstLatestRelease -->
```
BASE=$DEMO_HOME/base
mkdir $BASE

cat <<EOF >$BASE/kustomization.yaml
resources:
- pod.yaml
EOF

cat <<EOF >$BASE/pod.yaml
apiVersion: v1
kind: Pod
metadata:
  name: myapp-pod
  labels:
    app: myapp
spec:
  containers:
  - name: nginx
    image: nginx:1.7.9
EOF
```

Define a variant in namespace-a overlaying base:
<!-- @makeNamespaceA @testAgainstLatestRelease -->
```
NSA=$DEMO_HOME/namespace-a
mkdir $NSA

cat <<EOF >$NSA/kustomization.yaml
resources:
- namespace.yaml
- ../base
namespace: namespace-a
EOF

cat <<EOF >$NSA/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: namespace-a
EOF
```

Define a variant in namespace-b overlaying base:
<!-- @makeNamespaceB @testAgainstLatestRelease -->
```
NSB=$DEMO_HOME/namespace-b
mkdir $NSB

cat <<EOF >$NSB/kustomization.yaml
resources:
- namespace.yaml
- ../base
namespace: namespace-b
EOF

cat <<EOF >$NSB/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: namespace-b
EOF
```

Then define a _Kustomization_ composing two variants together:
<!-- @makeTopLayer @testAgainstLatestRelease -->
```
cat <<EOF >$DEMO_HOME/kustomization.yaml
resources:
- namespace-a
- namespace-b
EOF
```

Now the workspace has following directories
> ```
> .
> ├── base
> │   ├── kustomization.yaml
> │   └── pod.yaml
> ├── kustomization.yaml
> ├── namespace-a
> │   ├── kustomization.yaml
> │   └── namespace.yaml
> └── namespace-b
>     ├── kustomization.yaml
>     └── namespace.yaml
> ```

Confirm that the `kustomize build` output contains two pod objects from namespace-a and namespace-b.

<!-- @confirmVariants @testAgainstLatestRelease -->
```
test 2 == \
  $(kustomize build $DEMO_HOME| grep -B 4 "namespace: namespace-[ab]" | grep "name: myapp-pod" | wc -l); \
  echo $?  
```
