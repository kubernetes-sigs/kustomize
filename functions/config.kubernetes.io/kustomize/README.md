# Kustomize API

This is an alpha version of Kustomizations that are performed inline --
e.g. written back to the original input files.

This API is alpha and expected to change frequently.

## Example

Create this file

    # example/functions.yaml
    apiVersion: config.kubernetes.io/v1alpha1
    kind: InlineKustomization
    metadata:
      name: my-kustomization
      annotations:
        config.kubernetes.io/local-config: true # {"description": "do not apply"}
      configFn:
        container:
          image: gcr.io/kustomize-functions/apis:v0.0.1 # {"description": "kustomize config run ."}
    spec:
      commonLabels:
        foo: bar
      commonNamePrefix: bar-
      commonNamespace: default
      commonSelectors:
        foo: bar
      configMapGenerators:
      - name: cf
        literals:
        - a=b
        - c=d

Run `kustomize config run`

    $ kustomize config run example/
