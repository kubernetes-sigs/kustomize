## Images transformations

This tutorial shows how to modify images in resources, and create a custom images transformer configuration.

Create a workspace by
<!-- @createws @testAgainstLatestRelease -->
```
DEMO_HOME=$(mktemp -d)
```

### Adding a custom resource

Consider a Custom Resource Definition(CRD) of kind `MyKind` with field
- `.spec.runLatest.container.image` referencing an image

Add the following file to configure the images transformer for the CRD:

<!-- @addConfig @testAgainstLatestRelease -->
```
mkdir $DEMO_HOME/kustomizeconfig
cat > $DEMO_HOME/kustomizeconfig/mykind.yaml << EOF

images:
- path: spec/runLatest/container/image
  kind: MyKind
EOF
```

### Apply config

Create a file with some resources that includes an instance of `MyKind`:

<!-- @createResource @testAgainstLatestRelease -->
```
cat > $DEMO_HOME/resources.yaml << EOF

apiVersion: config/v1
kind: MyKind
metadata:
  name: testSvc
spec:
  runLatest:
    container:
      image: crd-image
  containers:
    - image: docker
      name: ecosystem
    - image: my-mysql
      name: testing-1
---
group: apps
apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      initContainers:
      - name: nginx2
        image: my-app
      - name: init-alpine
        image: alpine:1.8.0
EOF
```

Create a kustomization.yaml referring to it:

<!-- @createKustomization @testAgainstLatestRelease -->
```
cat > $DEMO_HOME/kustomization.yaml << EOF
resources:
- resources.yaml

images:
- name: crd-image
  newName: new-crd-image
  newTag: new-v1-tag
- name: my-app
  newName: new-app-1
  newTag: MYNEWTAG-1
- name: my-mysql
  newName: prod-mysql
  newTag: v3
- name: docker
  newName: my-docker2
  digest: sha256:25a0d4
EOF
```

Use the customized transformer configurations by specifying them
in the kustomization file:
<!-- @addTransformerConfigs @testAgainstLatestRelease -->
```
cat >> $DEMO_HOME/kustomization.yaml << EOF
configurations:
- kustomizeconfig/mykind.yaml
EOF
```

Run `kustomize build` and verify that the images have been updated.

<!-- @build @testAgainstLatestRelease -->
```
test 1 == \
$(kustomize build $DEMO_HOME | grep -A 2 ".*image" | grep "new-crd-image:new-v1-tag" | wc -l); \
echo $?
```

<!-- @build @testAgainstLatestRelease -->
```
test 1 == \
$(kustomize build $DEMO_HOME | grep -A 2 ".*image" | grep "new-app-1:MYNEWTAG-1" | wc -l); \
echo $?
```

<!-- @build @testAgainstLatestRelease -->
```
test 1 == \
$(kustomize build $DEMO_HOME | grep -A 2 ".*image" | grep "my-docker2@sha" | wc -l); \
echo $?
```
<!-- @build @testAgainstLatestRelease -->
```
test 1 == \
$(kustomize build $DEMO_HOME | grep -A 2 ".*image" | grep "prod-mysql:v3" | wc -l); \
echo $?
```