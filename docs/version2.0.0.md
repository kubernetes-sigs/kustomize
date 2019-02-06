# Kustomize 2.0.0

After security review, a field used in secret generation (see below) was removed from the definition of a kustomization file with no mechanism to convert it to a new form. Also, the set of files accessible from a kustomization file has been further constrained.

Per the [versioning policy](versioningPolicy.md), backward incompatible changes trigger an increment of the major version number, hence we go from 1.0.11 to 2.0.0. We're taking this major version increment opportunity to remove some already deprecated fields, and the code paths associated with them.

## Backward Incompatible Changes

### Kustomization Path Constraints
A kustomization file can specify paths to other files, including resources, patches, configmap generation data, secret generation data and bases. In the case of a base, the path can be a git URL instead.

In 1.x, these paths had to be relative to the current kustomization directory (the location of the kustomization file used in the `build` command).

In 2.0, bases can continue to specify, via relative paths, kustomizations outside the current kustomization directory.
But non-base paths are constrained to terminate in or below the current kustomization directory. Further, bases specified via a git URL may not reference files outside of the directory used to clone the repository.

### Kustomization Field Removals

#### patches
`patches` was deprecated and replaced by `patchesStrategicMerge` when `patchesJson6902` was introduced.
In Kustomize 2.0.0, `patches` is removed. Please use `patchesStrategicMerge` instead.

#### imageTags
`imageTags` is replaced by `images` since `images` can provide more features to change image names, registries, tags and digests.

#### secretGenerator/commands
`commands` is removed from SecretGenerator due to [security concern](https://docs.google.com/document/d/1FYgLVdq-siB_Cef9yuQBmit0PbrE8lsyTBdGI2eA2y8/edit). One can use `files` or `literals`, similar to ConfigMapGenerator, to generate a secret.
```
secretGenerator:
- name: app-tls
  files:
    - secret/tls.cert
    - secret/tls.key
  type: "kubernetes.io/tls"
```

## Compatible Changes (New Features)
As this release is triggered by a security change,
there are no major new features to announce. A few things that are worth mentioning in this release are:

* More than _40_ issues closed since 1.0.11 release (including many extensions to transformation rules).
* Users can run `kustomize edit fix` to migrate a kustomization file working with previous versions to one working with 2.0.0. For example, a kustomization.yaml with following content
  ```
  patches:
    - deployment-patch.yaml
  imageTags:
    - name: postgres
      newTag: v1
  ```
  
  will be converted to
  
  ```
  apiVersion: kustomize.config.k8s.io/v1beta1
  kind: Kustomization
  patchesStrategicMerge:
    - deployment-patch.yaml
  images:
    - name: postgres
      newTag: v1
  ```

* Kustomization filename

  In previous versions, the canonical name of a kustomization file is `kustomization.yaml`. Kustomize 2.0.0 is extended to recognize more file names: `kustomization.yaml`, `kustomization.yml` and `Kustomization`. In a directory, only one of those filenames is allowed. If there are more than one found, Kustomize will exit with an error. Please select the best filename for your use cases.
* No longer planning to deprecate namespace prefix/suffix. The deprecation warning
   ```
   Adding nameprefix and namesuffix to Namespace resource will be deprecated in next release.
   ```
    is removed. Since changing this behavior will break many users' workflow. Kustomize will continue with adding nameprefix and namesuffix to Namespace resources.
