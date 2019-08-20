# 例子: 应用 json patch（json补丁）

kustomization文件支持通过[JSON patches](https://tools.ietf.org/html/rfc6902)来修改已有的资源.

下面的例子将会使用这个功能对`Ingress`加以修改.

首先，创建一个包含`ingress`的`kustomization`文件.

<!-- @createIngress @testAgainstLatestRelease -->
```bash
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

定义一个JSON patch文件，以更新`Ingress`对象的2个字段:

- 把 host 从 `foo.bar.com` 改为 `foo.bar.io`
- 把 servicePort 从 `80` 改为 `8080`

<!-- @addJsonPatch @testAgainstLatestRelease -->
```bash
cat <<EOF >$DEMO_HOME/ingress_patch.json
[
  {"op": "replace", "path": "/spec/rules/0/host", "value": "foo.bar.io"},
  {"op": "replace", "path": "/spec/rules/0/http/paths/0/backend/servicePort", "value": 8080}
]
EOF
```

JSON patch 也可以写成 YAML 的格式.该例子顺便展示了“添加”操作：

<!-- @addYamlPatch @testAgainstLatestRelease -->
```bash
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

在kustomization.yaml文件中增加 _patchesJson6902_ 字段，以应用该补丁

<!-- @applyJsonPatch @testAgainstLatestRelease -->
```bash
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

运行 `kustomize build $DEMO_HOME`, 在输出那里确认 host 已经被正确更新.

<!-- @confirmHost @testAgainstLatestRelease -->
```bash
test 1 == \
  $(kustomize build $DEMO_HOME | grep "host: foo.bar.io" | wc -l); \
  echo $?
```

运行 `kustomize build $DEMO_HOME`, 在输出那里确认 servicePort 已经被正确更新.

<!-- @confirmServicePort @testAgainstLatestRelease -->

```bash
test 1 == \
  $(kustomize build $DEMO_HOME | grep "servicePort: 8080" | wc -l); \
  echo $?
```

如果 patch 是YAML格式的，就能正确解析:

<!-- @applyYamlPatch @testAgainstLatestRelease -->
```bash
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

运行 `kustomize build $DEMO_HOME`, 在输出那里确认有 `/test` 这个路径.

<!-- @confirmYamlPatch @testAgainstLatestRelease -->
```bash
test 1 == \
  $(kustomize build $DEMO_HOME | grep "path: /test" | wc -l); \
  echo $?
```