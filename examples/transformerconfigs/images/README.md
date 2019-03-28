## Images transformations

This tutorial shows how to modify images in resources, and create custom images transformer configurations.

Create a workspace by
<!-- @createws @test -->
```
DEMO_HOME=$(mktemp -d)
```

### Adding a custom resource

Consider a Custom Resource Definition(CRD) of kind `MyKind` with field
- `.spec.runLatest.configuration.revisionTemplate.spec.container.image` referencing an image

Add the following file to configure the images transformer for the CRD:

<!-- @addConfig @test -->
```
mkdir $DEMO_HOME/kustomizeconfig
cat > $DEMO_HOME/kustomizeconfig/mykind.yaml << EOF

images:
- path: spec/runLatest/configuration/revisionTemplate/spec/container/image
  kind: MyKind
EOF
```

### Apply config

Create a file with some resources that includes an instance of `MyKind`:

<!-- @createResource @test -->
```
cat > $DEMO_HOME/resources.yaml << EOF

apiVersion: mykind/config/v1
kind: MyKind
metadata:
  name: testSvc
spec:
  runLatest:
    configuration:
      revisionTemplate:
        spec:
          container:
            - image: my-app
              name: my-app-name
  containers:
    - image: docker
      name: ecosystem
    - image: my-mysql
      name: solar
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

Create a kustomization referring to it:

<!-- @createKustomization @test -->
```
cat > $DEMO_HOME/kustomization.yaml << EOF
resources:
- resources.yaml

configurations:
- mykind.yaml

images:
- name: my-app
  newName: bear1
  newTag: MYNEWTAG1
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
<!-- @addTransformerConfigs @test -->
```
cat >> $DEMO_HOME/kustomization.yaml << EOF
configurations:
- kustomizeconfig/mykind.yaml
EOF
```

Run `kustomize build` and verify that the list of images have been updated.

<!-- TODO, fixme, @build @test -->
```
test 2 == \
$(kustomize build $DEMO_HOME | grep -A 2 ".*image" | grep "bear1" | wc -l); \
echo $?
```