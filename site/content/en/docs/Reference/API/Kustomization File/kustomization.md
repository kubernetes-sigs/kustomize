---
title: "kustomization"
linkTitle: "kustomization"
type: docs
weight: 1
description: >
    Kustomization contains information to generate customized resources.
---

`apiVersion: kustomize.config.k8s.io/v1beta1`

### Kustomization

---

* **apiVersion**: kustomize.config.k8s.io/v1beta1
* **kind**: Kustomization
* **openAPI** (map[string]string)

	[OpenAPI]({{< ref "openapi.md" >}}) contains information about what kubernetes schema to use.

* **namePrefix** (string)

	[NamePrefix]({{< ref "namePrefix.md" >}}) will prefix the names of all resources mentioned in the kustomization file including generated configmaps and secrets.

* **nameSuffix** (string)

	[NameSuffix]({{< ref "nameSuffix.md" >}}) will suffix the names of all resources mentioned in the kustomization file including generated configmaps and secrets.

* **namespace** (string)

	[Namespace]({{< ref "namespace.md" >}}) to add to all objects.

* **commonLabels** (map[string]string)

	[CommonLabels]({{< ref "commonLabels.md" >}}) to add to all objects and selectors.

* **labels** ([][Label]({{< ref "labels.md" >}}))

	Labels to add to all objects but not selectors.

* **commonAnnotations** (map[string]string)

	[CommonAnnotations]({{< ref "commonAnnotations.md" >}}) to add to all objects.

* **patchesStrategicMerge** ([][PatchStrategicMerge]({{< ref "patchesStrategicMerge.md" >}}))

	Deprecated: Use the Patches field instead, which provides a superset of the functionality of PatchesStrategicMerge.

* **patchesJson6902** ([][Patch]({{< ref "patches.md" >}}))

	Deprecated: Use the Patches field instead, which provides a superset of the functionality of JSONPatches. [JSONPatches]({{< ref "patchesjson6902.md" >}}) is a list of JSONPatch for applying JSON patch.

* **patches** ([][Patch]({{< ref "patches.md" >}}))

	Patches is a list of patches, where each one can be either a Strategic Merge Patch or a JSON patch. Each patch can be applied to multiple target objects.

* **images** ([][Image]({{< ref "images.md" >}}))

	Images is a list of (image name, new name, new tag or digest) for changing image names, tags or digests. This can also be achieved with a patch, but this operator is simpler to specify.

* **imageTags** ([][Image]({{< ref "images.md" >}}))

	Deprecated: Use the Images field instead.

* **replacements** ([][ReplacementField]({{< ref "replacements.md" >}}))

	Replacements is a list of replacements, which will copy nodes from a specified source to N specified targets.

* **replicas** ([][Replica]({{< ref "replicas.md" >}}))

	Replicas is a list of {resourcename, count} that allows for simpler replica specification. This can also be done with a patch.

* **vars** ([][Var]({{< ref "vars.md" >}}))

	Deprecated: Vars will be removed in future release. Migrate to Replacements instead. Vars allow things modified by kustomize to be injected into a kubernetes object specification.

* **sortOptions** ([sortOptions]({{< ref "sortOptions.md" >}}))

	SortOptions change the order that kustomize outputs resources.

* **resources** ([]string)

	[Resources]({{< ref "resources.md" >}}) specifies relative paths to files holding YAML representations of kubernetes API objects, or specifications of other kustomizations via relative paths, absolute paths, or URLs.

* **components** ([]string)

	[Components]({{< ref "components.md" >}}) specifies relative paths to specifications of other Components via relative paths, absolute paths, or URLs.

* **crds** ([]string)

	[Crds]({{< ref "crds.md" >}}) specifies relative paths to Custom Resource Definition files. This allows custom resources to be recognized as operands, making it possible to add them to the Resources list. CRDs themselves are not modified.

* **bases** ([]string)

	Deprecated: Anything that would have been specified here should be specified in the Resources field instead. [Bases]({{< ref "bases.md" >}}) specifies relative paths to files holding YAML representations of Kubernetes API objects.

* **configMapGenerator** ([][ConfigMapArgs]({{< ref "configMapGenerator.md#configmapargs" >}}))

	[ConfigMapGenerator]({{< ref "configMapGenerator.md" >}}) is a list of configmaps to generate from local data (one configMap per list item). The resulting resource is a normal operand, subject to name prefixing, patching, etc.  By default, the name of the map will have a suffix hash generated from its contents.

* **secretGenerator** ([][SecretArgs]({{< ref "secretGenerator.md#secretargs" >}}))

	[SecretGenerator]({{< ref "secretGenerator.md" >}}) is a list of secrets to generate from local data (one secret per list item). The resulting resource is a normal operand, subject to name prefixing, patching, etc.  By default, the name of the map will have a suffix hash generated from its contents.

* **helmGlobals** (HelmGlobals)

	HelmGlobals contains helm configuration that isn't chart specific.

* **helmCharts** ([][HelmChart]({{< ref "helmCharts.md" >}}))

	HelmCharts is a list of helm chart configuration instances.

* **helmChartInflationGenerator** ([]HelmChartArgs)

	Deprecated: Auto-converted to HelmGlobals and [HelmCharts]({{< ref "helmCharts.md" >}}). HelmChartInflationGenerator is a list of helm chart configurations.

* **generatorOptions** ([GeneratorOptions]({{< ref "generatorOptions.md" >}}))

	GeneratorOptions modify behavior of all ConfigMap and Secret generators.

* **configurations** ([]string)

	Configurations is a list of transformer configuration files

* **generators** ([]string)

	Generators is a list of files containing custom generators

* **transformers** ([]string)

	Transformers is a list of files containing transformers

* **validators** ([]string)

	Validators is a list of files containing validators

* **buildMetadata** ([]string)

	[BuildMetadata]({{< ref "buildMetadata.md" >}}) is a list of strings used to toggle different build options
