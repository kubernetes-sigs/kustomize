# JSON Patching

[JSON patches]: https://tools.ietf.org/html/rfc6902
[JSON patch]: https://tools.ietf.org/html/rfc6902

A kustomization file supports customizing
resources via [JSON patches].

Make a place to work:

<!-- @placeToWork @testAgainstLatestRelease -->
```
DEMO_HOME=$(mktemp -d)
```

We'll be editting an `Ingress` object:

<!-- @ingress @testAgainstLatestRelease -->
```
cat <<EOF >$DEMO_HOME/ingress.yaml
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: my-ingress
spec:
  rules:
  - host: foo.bar.com
    http:
      paths:
      - path: /
        backend:
          serviceName: homepage
          servicePort: 8888
      - path: /api
        backend:
          serviceName: my-api
          servicePort: 7701
      - path: /test
        backend:
          serviceName: hello
          servicePort: 7702
EOF
```

The edits we want to make are:

 - change the value of `host` to _foo.bar.io_
 - change the port for `'/'` from _8888_ to _80_
 - insert an entirely new serving path `/healthz`
   at a particular point in the `paths` list,
   rather than at the end or the beginning.

Here's the patch file to do that:

<!-- @addJsonPatch @testAgainstLatestRelease -->
```
cat <<EOF >$DEMO_HOME/ingress_patch.json
[
  {"op": "replace",
   "path": "/spec/rules/0/host",
   "value": "foo.bar.io"},

  {"op": "replace",
   "path": "/spec/rules/0/http/paths/0/backend/servicePort",
   "value": 80},

  {"op": "add",
   "path": "/spec/rules/0/http/paths/1",
   "value": { "path": "/healthz", "backend": {"servicePort":7700} }}
]
EOF
```

We'll of course need a `kustomization` file
referring to the `Ingress`:

<!-- @kustomization @testAgainstLatestRelease -->
```
cat <<EOF >$DEMO_HOME/kustomization.yaml
resources:
- ingress.yaml
EOF
```

To this same `kustomization` file, add a
`patches` field refering to
the patch file we just made and
target it to the `Ingress` object:

<!-- @applyJsonPatch @testAgainstLatestRelease -->
```
cat <<EOF >>$DEMO_HOME/kustomization.yaml
patches:
- path: ingress_patch.json
  target:
    group: networking.k8s.io
    version: v1beta1
    kind: Ingress
    name: my-ingress
EOF
```

Define the expected output:
<!-- @expected @testAgainstLatestRelease -->
```
cat <<EOF >$DEMO_HOME/out_expected.yaml
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: my-ingress
spec:
  rules:
  - host: foo.bar.io
    http:
      paths:
      - backend:
          serviceName: homepage
          servicePort: 80
        path: /
      - backend:
          servicePort: 7700
        path: /healthz
      - backend:
          serviceName: my-api
          servicePort: 7701
        path: /api
      - backend:
          serviceName: hello
          servicePort: 7702
        path: /test
EOF
```

Run the build:
<!-- @runIt @testAgainstLatestRelease -->
```
kustomize build $DEMO_HOME >$DEMO_HOME/out_actual.yaml
```

Confirm they match:

<!-- @diffShouldExitZero @testAgainstLatestRelease -->
```
diff $DEMO_HOME/out_actual.yaml $DEMO_HOME/out_expected.yaml
```

If you prefer YAML to JSON, the patch can be expressed
in YAML format (nevertheless following [JSON patch] rules):

<!-- @writeYamlPatch @testAgainstLatestRelease -->
```
cat <<EOF >$DEMO_HOME/ingress_patch.yaml
- op: add
  path: /spec/rules/0/http/paths/-
  value:
    path: '/canada'
    backend:
      serviceName: hoser
      servicePort: 7703
EOF
```

Now add this to the list of patches in the `kustomization` file:

<!-- @addYamlPatch @testAgainstLatestRelease -->
```
cat <<EOF >>$DEMO_HOME/kustomization.yaml
- path: ingress_patch.yaml
  target:
    group: networking.k8s.io
    version: v1beta1
    kind: Ingress
    name: my-ingress
EOF
```

We expect the following at the end of the output:
<!-- @expected @testAgainstLatestRelease -->
```
cat <<EOF >$DEMO_HOME/out_expected.yaml
      - backend:
          serviceName: hello
          servicePort: 7702
        path: /test
      - backend:
          serviceName: hoser
          servicePort: 7703
        path: /canada
EOF
```

Try it:

<!-- @runIt @testAgainstLatestRelease -->
```
kustomize build $DEMO_HOME | tail -n 8 |\
    diff  $DEMO_HOME/out_expected.yaml -
```

To see how to apply one JSON patch to many resources,
see the [multi-patch](patchMultipleObjects.md) demo.
