# Helm Namespace Example

This example exercises namespace transformation on resources generated from a local Helm chart.

The example kustomization sets a namespace for Helm-generated resources that do not already declare one, while preserving any namespace explicitly emitted by the chart.

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
  --helm-command "$helmCommand" \
  "$EXAMPLE_HOME")
printf '%s\n' "$output"
```

The Service is emitted by the chart without a namespace, so the example namespace should be applied:

<!-- @checkMissingNamespaceFilled @testHelm -->
```sh
printf '%s\n' "$output" | grep -A4 'name: chart-service' | grep 'namespace: top-level-ns'
```

The ConfigMap is emitted by the chart with an explicit namespace, so that value should be preserved:

<!-- @checkExistingNamespacePreserved @testHelm -->
```sh
printf '%s\n' "$output" | grep -A4 'name: chart-config' | grep 'namespace: chart-owned-ns'
```

The Secret is emitted by the chart with an release namespace, so that value should be preserved:

<!-- @checkExistingNamespacePreserved @testHelm -->
```sh
printf '%s\n' "$output" | grep -A4 'name: chart-secret' | grep 'namespace: chart-ns'
```
