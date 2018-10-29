## Transformer Configurations - labels

This tutorial shows how to disable adding common labels to fields in some kind of resources.

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
To stop adding common labels to a field in some kind of resources. all need to do is to remove the corresponding configuration for that field from `commonlabels.yaml` file. For example, remove following block from this file will prevent Kustomize from adding labels to field `spec/egress/from/podSelector/matchLabels` in NetworkPolicy resources.
```
- path: spec/ingress/from/podSelector/matchLabels
  create: false
  group: networking.k8s.io
  kind: NetworkPolicy
```
<!-- @removeConfig @test -->
```
sed -i -e '/- path: spec\/ingress\/from\/podSelector\/matchLabels/,+3d' ~/.kustomize/config/commonlabels.yaml
```

### Apply config
Create a kustomization with a NetworkPolicy instance.

<!-- @createKustomization @test -->
```
DEMO_HOME=$(mktemp -d)

cat > $DEMO_HOME/kustomization.yaml << EOF
resources:
- resources.yaml

commonLabels:
  foo: bar
EOF

cat > $DEMO_HOME/resources.yaml << EOF
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: test-network-policy
spec:
  podSelector:
    matchLabels:
      role: db
  policyTypes:
  - Ingress
  ingress:
  - from:
    - ipBlock:
        cidr: 172.17.0.0/16
        except:
        - 172.17.1.0/24
    - podSelector:
        matchLabels:
          role: frontend
    ports:
    - protocol: TCP
      port: 6379
EOF
```

Run `kustomize build` with customized transformer configurations and verify that
`foo: bar` is not added to `spec/ingress/from/podSelector/matchLabels`.

<!-- @build @test -->
```
test 0 == \
$(kustomize build $DEMO_HOME -t ~/.kustomize/config | grep -A 3 "\- podSelector:" | grep "foo: bar" | wc -l); \
echo $?    
```
