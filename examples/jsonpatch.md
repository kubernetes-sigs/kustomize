# Demo: applying a json patch

A kustomization file supports customizing resources via [JSON patches](https://tools.ietf.org/html/rfc6902).

The example below modifies an `Ingress` object with such a patch.

Make a `kustomization` containing an ingress resource.

<!-- @createIngress @testAgainstLatestRelease -->
```
DEMO_HOME=$(mktemp -d)

cat <<EOF >$DEMO_HOME/kustomization.yaml
resources:
- ingress.yaml
EOF

cat <<EOF >$DEMO_HOME/ingress.yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: my-ingress
spec:
  rules:
  - host: foo.bar.com
    http:
      paths:
      - backend:
          serviceName: my-api
          servicePort: 80
EOF
```

Declare a JSON patch file to update two fields of the Ingress object:

- change host from `foo.bar.com` to `foo.bar.io`
- change servicePort from `80` to `8080`

<!-- @addJsonPatch @testAgainstLatestRelease -->
```
cat <<EOF >$DEMO_HOME/ingress_patch.json
[
  {"op": "replace", "path": "/spec/rules/0/host", "value": "foo.bar.io"},
  {"op": "replace", "path": "/spec/rules/0/http/paths/0/backend/servicePort", "value": 8080}
]
EOF
```

You can also write the patch in YAML format. This example also shows the "add" operation:

<!-- @addYamlPatch @testAgainstLatestRelease -->
```
cat <<EOF >$DEMO_HOME/ingress_patch.yaml
- op: replace
  path: /spec/rules/0/host
  value: foo.bar.io

- op: add
  path: /spec/rules/0/http/paths/-
  value:
    path: '/test'
    backend:
      serviceName: my-test
      servicePort: 8081
EOF
```

Apply the patch by adding _patchesJson6902_ field in kustomization.yaml

<!-- @applyJsonPatch @testAgainstLatestRelease -->
```
cat <<EOF >>$DEMO_HOME/kustomization.yaml
patchesJson6902:
- target:
    group: extensions
    version: v1beta1
    kind: Ingress
    name: my-ingress
  path: ingress_patch.json
EOF
```

Running `kustomize build $DEMO_HOME`, in the output confirm that host has been updated correctly.
<!-- @confirmHost @testAgainstLatestRelease -->
```
test 1 == \
  $(kustomize build $DEMO_HOME | grep "host: foo.bar.io" | wc -l); \
  echo $?
```
Running `kustomize build $DEMO_HOME`, in the output confirm that the servicePort has been updated correctly.
<!-- @confirmServicePort @testAgainstLatestRelease -->
```
test 1 == \
  $(kustomize build $DEMO_HOME | grep "servicePort: 8080" | wc -l); \
  echo $?
```

If the patch is YAML-formatted, it will be parsed correctly:

<!-- @applyYamlPatch @testAgainstLatestRelease -->
```
cat <<EOF >>$DEMO_HOME/kustomization.yaml
patchesJson6902:
- target:
    group: extensions
    version: v1beta1
    kind: Ingress
    name: my-ingress
  path: ingress_patch.yaml
EOF
```

<!-- @confirmYamlPatch @testAgainstLatestRelease -->
```
test 1 == \
  $(kustomize build $DEMO_HOME | grep "path: /test" | wc -l); \
  echo $?
```
