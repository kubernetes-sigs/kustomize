## Transformer Configurations - CRD

This tutorial shows how to add transformer configurations to support a CRD type.

### Get Default Config
Get the default transformer configurations by

<!-- @saveConfig @test -->
```
kustomize config save -d ~/.kustomize/config
```
The default configurations are save in directory `~/.kustomize/config` as several files

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
cat > ~/.kustomize/config/mykind.yaml << EOF

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
DEMO_HOME=$(mktemp -d)

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

Run `kustomize build` with customized transformer configurations and verify that
the namereference is correctly resolved.

<!-- @build @test -->
```
test 2 == \
$(kustomize build $DEMO_HOME -t ~/.kustomize/config | grep -A 2 ".*Ref" | grep "test-" | wc -l); \
echo $?    
```

Run `kustomize build` with customized transformer configurations and verify that
the vars correctly resolved.

<!-- @verify @test -->
```
test 0 == \
$(kustomize build $DEMO_HOME -t ~/.kustomize/config | grep "BEE_ACTION" | wc -l); \
echo $?    
```

To understand this better, compare the output using default transformer configurations.

<!-- @compareOutput -->
```
diff \
 <(kustomize build $DEMO_HOME) \
 <(kustomize build $DEMO_HOME -t ~/.kustomize/config ) |\
 more
```

The difference output should look something like
> ```
> 20,21c20,21
> <     action: $(BEE_ACTION)
> <     name: bee
> ---
> >     action: fly
> >     name: test-bee
> 25c25
> <     - $(BEE_ACTION)
> ---
> >     - fly
> 28c28,30
> <     name: crdsecret
> ---
> >     name: test-crdsecret
> >   selectors:
> >     foo: bar
> ```


