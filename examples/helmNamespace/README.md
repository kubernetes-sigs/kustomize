# Helm Namespace Example

This example exercises namespace transformation on resources generated from a local Helm chart.

A namespace is set for Helm-generated resources that do not already declare one, while preserving any namespace explicitly emitted by the chart.

Two kustomizations share the single chart in [charts](charts):

* [topLevelNamespace](topLevelNamespace/kustomization.yaml) sets a top-level `namespace`.
* [chartOnly](chartOnly/kustomization.yaml) sets no top-level `namespace`; the namespace comes only from `helmCharts[].namespace`.

Because the shared chart lives above each kustomization root, the builds need `--load-restrictor LoadRestrictionsNone`.

## Build the example

This example defines the `helm` command as:

<!-- @defineHelmCommand @testHelm -->
```sh
helmCommand=${MYGOBIN:-~/go/bin}/helmV3
```

Use the checked-in example:

<!-- @defineExampleHome @testHelm -->
```sh
EXAMPLE_HOME=examples/helmNamespace
```

Build the example with Helm enabled:

<!-- @buildOverlay @testHelm -->
```sh
output=$(kustomize build \
  --enable-helm \
  --load-restrictor LoadRestrictionsNone \
  --helm-command "$helmCommand" \
  "$EXAMPLE_HOME/topLevelNamespace")
printf '%s\n' "$output"
```

## Helm Chart with Namespace in `helmChart.namespace`

The Service is emitted by the chart without a namespace, so the example namespace should be applied:

<!-- @checkMissingNamespaceFilled @testHelm -->
```sh
printf '%s\n' "$output" | grep -A4 'name: test-a-service' | grep 'namespace: chart-ns'
```

The ConfigMap is emitted by the chart with an explicit namespace, so that value should be preserved:

<!-- @checkExistingNamespacePreserved @testHelm -->
```sh
printf '%s\n' "$output" | grep -A4 'name: test-a-config' | grep 'namespace: chart-owned-ns'
```

The Secret is emitted by the chart with an release namespace, so that value should be preserved:

<!-- @checkExistingNamespacePreserved @testHelm -->
```sh
printf '%s\n' "$output" | grep -A4 'name: test-a-secret' | grep 'namespace: chart-ns'
```

## Helm Chart without Namespace in `helmChart.namespace`

The Service is emitted by the chart without a namespace, so the example namespace should be applied:

<!-- @checkMissingNamespaceFilled @testHelm -->
```sh
printf '%s\n' "$output" | grep -A4 'name: test-b-service' | grep 'namespace: top-level-ns'
```

The ConfigMap is emitted by the chart with an explicit namespace, so that value should be preserved:

<!-- @checkExistingNamespacePreserved @testHelm -->
```sh
printf '%s\n' "$output" | grep -A4 'name: test-b-config' | grep 'namespace: chart-owned-ns'
```

The Secret is emitted by the chart with an release namespace, so that value should be preserved:

<!-- @checkExistingNamespacePreserved @testHelm -->
```sh
printf '%s\n' "$output" | grep -A4 'name: test-b-secret' | grep 'namespace: top-level-ns'
```

## RoleBinding subjects

The chart emits a RoleBinding with three subjects. A subject named `default`
without a namespace is filled with the effective namespace, just like any other
missing namespace field:

<!-- @checkRoleBindingSubjectFilled @testHelm -->
```sh
printf '%s\n' "$output" | grep -A14 'name: test-a-rolebinding' | grep -A3 'subjects:' | grep 'namespace: chart-ns'
```

<!-- @checkRoleBindingSubjectFilled @testHelm -->
```sh
printf '%s\n' "$output" | grep -A14 'name: test-b-rolebinding' | grep -A3 'subjects:' | grep 'namespace: top-level-ns'
```

A subject namespace rendered by the chart is preserved:

<!-- @checkRoleBindingSubjectPreserved @testHelm -->
```sh
printf '%s\n' "$output" | grep -A14 'name: test-a-rolebinding' | grep 'namespace: chart-owned-ns'
```

<!-- @checkRoleBindingSubjectPreserved @testHelm -->
```sh
printf '%s\n' "$output" | grep -A14 'name: test-b-rolebinding' | grep 'namespace: chart-owned-ns'
```

Subjects not named `default` are never touched, in any release:

<!-- @checkRoleBindingOtherSubjectUntouched @testHelm -->
```sh
! printf '%s\n' "$output" | grep -A1 'name: app-service-account' | grep 'namespace:'
```

## Namespace only in `helmChart.namespace`, no top-level `namespace`

The [chartOnly](chartOnly/kustomization.yaml) kustomization sets no `namespace`
at all, so no explicit namespace transformer is configured:

<!-- @buildChartOnly @testHelm -->
```sh
outputChartOnly=$(kustomize build \
  --enable-helm \
  --load-restrictor LoadRestrictionsNone \
  --helm-command "$helmCommand" \
  "$EXAMPLE_HOME/chartOnly")
printf '%s\n' "$outputChartOnly"
```

The Service is emitted by the chart without a namespace, so the `helmCharts[].namespace` value should be applied:

<!-- @checkMissingNamespaceFilledChartOnly @testHelm -->
```sh
printf '%s\n' "$outputChartOnly" | grep -A4 'name: test-c-service' | grep 'namespace: helm-only-ns'
```

The ConfigMap is emitted by the chart with an explicit namespace, so that value should be preserved:

<!-- @checkExistingNamespacePreservedChartOnly @testHelm -->
```sh
printf '%s\n' "$outputChartOnly" | grep -A4 'name: test-c-config' | grep 'namespace: chart-owned-ns'
```

The Secret is emitted by the chart with a release namespace, so that value should be preserved:

<!-- @checkReleaseNamespacePreservedChartOnly @testHelm -->
```sh
printf '%s\n' "$outputChartOnly" | grep -A4 'name: test-c-secret' | grep 'namespace: helm-only-ns'
```

The RoleBinding subject named `default` without a namespace is filled from
`helmCharts[].namespace`, while the subject namespace rendered by the chart is
preserved:

<!-- @checkRoleBindingSubjectFilledChartOnly @testHelm -->
```sh
printf '%s\n' "$outputChartOnly" | grep -A14 'name: test-c-rolebinding' | grep -A3 'subjects:' | grep 'namespace: helm-only-ns'
```

<!-- @checkRoleBindingSubjectPreservedChartOnly @testHelm -->
```sh
printf '%s\n' "$outputChartOnly" | grep -A14 'name: test-c-rolebinding' | grep 'namespace: chart-owned-ns'
```

## Generated ConfigMap and Secret references in a Helm Deployment

The chart also emits a Deployment that mounts the kustomize-generated `mounted-config` ConfigMap and `mounted-secret` Secret as volumes. The generators add a name-suffix hash, so the Deployment's volume references must keep that hash. This guards against [issue #6077](https://github.com/kubernetes-sigs/kustomize/issues/6077), where setting a namespace dropped the hash from references in Helm-generated workloads.

The namespace is applied to the Helm-generated Deployment, just like the other resources:

<!-- @checkDeploymentNamespaceFilled @testHelm -->
```sh
printf '%s\n' "$output" | grep -A2 'name: test-a-deployment' | grep 'namespace: chart-ns'
```

<!-- @checkDeploymentNamespaceFilled @testHelm -->
```sh
printf '%s\n' "$output" | grep -A2 'name: test-b-deployment' | grep 'namespace: top-level-ns'
```

The `test-b` Deployment shares the `top-level-ns` namespace with the generated ConfigMap and Secret, so its volume references resolve to the full hashed names:

<!-- @checkHashedConfigMapReference @testHelm -->
```sh
printf '%s\n' "$output" | grep -A40 'name: test-b-deployment' | grep -A1 'configMap:' | grep 'name: mounted-config-g46hh6k8tf'
```

<!-- @checkHashedSecretReference @testHelm -->
```sh
printf '%s\n' "$output" | grep -A40 'name: test-b-deployment' | grep 'secretName: mounted-secret-gh24bh7t8g'
```
