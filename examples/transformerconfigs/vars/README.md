## Transformer Configurations - vars

This tutorial shows how to add extra fields for variable substitution transformer.

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

### Update Config
To support extra field in variable substitution transformer, all need to do is to add the corresponding configuration in `varreference.yaml`. For example, adding following blocks enables variable substitution in secretConfigs inside annotations for Service objects.
<!-- @addConfig @test -->
```
cat >> ~/.kustomize/config/varreference.yaml << EOF
- path: metadata/annotations/secretConfigs
  kind: Service
EOF
```

### Apply config
Create a kustomization with a var:

<!-- @createKustomization @test -->
```
DEMO_HOME=$(mktemp -d)

cat > $DEMO_HOME/kustomization.yaml << EOF
resources:
- resource.yaml

vars:
- name: MY_SECRET
  objref:
    kind: Secret
    name: mysecret
    apiVersion: v1

nameprefix: demo-
 
EOF

cat > $DEMO_HOME/resource.yaml << EOF
apiVersion: v1
kind: Service
metadata:
  name: myservice
  annotations:
    secretConfigs: |
      ---
      secretName: \$(MY_SECRET)
---
apiVersion: v1
kind: Secret
metadata:
  name: mysecret
type: Opaque
data:
  username: YWRtaW4=
  password: MWYyZDFlMmU2N2Rm
EOF
```

Run `kustomize build` with customized transformer configurations and verify that
`$(MY_SECRET)` is replaced correctly by `demo-mysecret`.

<!-- @build @test -->
```
test 1 == \
$(kustomize build $DEMO_HOME -t ~/.kustomize/config | grep -A 3 "annotations" | grep "secretName: demo-mysecret" | wc -l); \
echo $?
```
