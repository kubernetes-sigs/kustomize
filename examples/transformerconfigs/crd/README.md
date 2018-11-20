## Transformer Configurations - CRD

This tutorial shows how to add transformer configurations to support a CRD type.

Create a workspace by
<!-- @createws @test -->
```
DEMO_HOME=$(mktemp -d)
```

### Get Default Config
Get the default transformer configurations by

<!-- @saveConfig @test -->
```
kustomize config save -d $DEMO_HOME/kustomizeconfig
```
The default configurations are save in directory `$DEMO_HOME/kusotmizeconfig` as several files

> ```
>  commonannotations.yaml  commonlabels.yaml  nameprefix.yaml  namereference.yaml  namespace.yaml  varreference.yaml
> ```

### Add Config for a CRD
All transformers will be involved for a CRD type. The default configurations already include some common fieldSpec for all types:

- nameprefix is added to `.metadata.name`
- namespace is added to `.metadata.namespace`
- labels is added to `.metadata.labels`
- annotations is added to `.metadata.annotations`

Thus those fieldSpec don't need to be added to support a CRD type.
Consider a CRD type `MyKind` with fields
- `.spec.secretRef.name` reference a Secret
- `.spec.beeRef.name` reference an instance of CRD `Bee`
- `.spec.containers.command` as the list of container commands
- `.spec.selectors` as the label selectors

Add following file to configure the transformers for the above fields
<!-- @addConfig @test -->
```
cat > $DEMO_HOME/kustomizeconfig/mykind.yaml << EOF

commonLabels:
- path: spec/selectors
  create: true
  kind: MyKind

nameReference:
- kind: Bee
  fieldSpecs:
  - path: spec/beeRef/name
    kind: MyKind
- kind: Secret
  fieldSpecs:
  - path: spec/secretRef/name
    kind: MyKind

varReference:
- path: spec/containers/command
  kind: MyKind
- path: spec/beeRef/action
  kind: MyKind

EOF
```

### Apply config
Create a kustomization with a `MyKind` instance.

<!-- @createKustomization @test -->
```
cat > $DEMO_HOME/kustomization.yaml << EOF
resources:
- resources.yaml

namePrefix: test-

commonLabels:
  foo: bar

vars:
- name: BEE_ACTION
  objref:
    kind: Bee
    name: bee
    apiVersion: v1beta1
  fieldref:
    fieldpath: spec.action
EOF

cat > $DEMO_HOME/resources.yaml << EOF
apiVersion: v1
kind: Secret
metadata:
  name: crdsecret
data:
  PATH: YmJiYmJiYmIK
---
apiVersion: v1beta1
kind: Bee
metadata:
  name: bee
spec:
  action: fly
---
apiVersion: jingfang.example.com/v1beta1
kind: MyKind
metadata:
  name: mykind
spec:
  secretRef:
    name: crdsecret
  beeRef:
    name: bee
    action: \$(BEE_ACTION)
  containers:
  - command:
    - "echo"
    - "\$(BEE_ACTION)"
    image: myapp
EOF
```

Use the cusotmized transformer configurations by adding those files in kustomization.yaml
<!-- @addTransformerConfigs @test -->
```
cat >> $DEMO_HOME/kustomization.yaml << EOF
configurations:
- kustomizeconfig/mykind.yaml
- kustomizeconfig/commonannotations.yaml
- kustomizeconfig/commonlabels.yaml
- kustomizeconfig/nameprefix.yaml
- kustomizeconfig/namereference.yaml
- kustomizeconfig/namespace.yaml
- kustomizeconfig/varreference.yaml
EOF
```

Run `kustomize build` and verify that the namereference is correctly resolved.

<!-- @build @test -->
```
test 2 == \
$(kustomize build $DEMO_HOME | grep -A 2 ".*Ref" | grep "test-" | wc -l); \
echo $?    
```

Run `kustomize build` and verify that the vars correctly resolved.

<!-- @verify @test -->
```
test 0 == \
$(kustomize build $DEMO_HOME | grep "BEE_ACTION" | wc -l); \
echo $?    
```
