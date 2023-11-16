---
title: "kustomization"
linkTitle: "kustomization"
type: docs
weight: 1
description: >
    Kustomiation contains information to generate customized resources.
---

`apiVersion: kustomize.config.k8s.io/v1beta1`

### Kustomization

---

* **apiVersion**: kustomize.config.k8s.io/v1beta1
* **kind**: Kustomization
* **OpenAPI** (map[string]string)

	OpenAPI contains information about what kubernetes schema to use.

* **NamePrefix** (string)

	NamePrefix will prefix the names of all resources mentioned in the kustomization file including generated configmaps and secrets.

* **NameSuffix** (string)

	NameSuffix will suffix the names of all resources mentioned in the kustomization file including generated configmaps and secrets.

* **Namespace** (string)

	Namespace to add to all objects.

* **CommonLabels** (map[string]string)

	CommonLabels to add to all objects and selectors.

* **Labels** ([]Label)

	Labels to add to all objects but not selectors.

* **commonAnnotations** (map[string]string)

	CommonAnnotations to add to all objects.

* **PatchesStrategicMerge** ([]PatchStrategicMerge)

	Deprecated: Use the Patches field instead, which provides a superset of the functionality of PatchesStrategicMerge.

* **PatchesJson6902** ([]Patch)

	Deprecated: Use the Patches field instead, which provides a superset of the functionality of JSONPatches. JSONPatches is a list of JSONPatch for applying JSON patch.

* **Patches** ([]Patch)

	Patches is a list of patches, where each one can be either a Strategic Merge Patch or a JSON patch. Each patch can be applied to multiple target objects.

* **Images** ([]Image)

	Images is a list of (image name, new name, new tag or digest) for changing image names, tags or digests. This can also be achieved with a patch, but this operator is simpler to specify.

* **ImageTags** ([]Image)

	Deprecated: Use the Images field instead.

* **Replacements** ([]ReplacementField)

	Replacements is a list of replacements, which will copy nodes from a specified source to N specified targets.

* **Replicas** ([]Replica)

	Replicas is a list of {resourcename, count} that allows for simpler replica specification. This can also be done with a patch.

* **Vars** ([]Var)

	Deprecated: Vars will be removed in future release. Migrate to Replacements instead. Vars allow things modified by kustomize to be injected into a kubernetes object specification.

* **SortOptions** (SortOptions)

	SortOptions change the order that kustomize outputs resources.

* **Resources** ([]string)

	Resources specifies relative paths to files holding YAML representations of kubernetes API objects, or specifications of other kustomizations via relative paths, absolute paths, or URLs.

* **Components** ([]string)

	Components specifies relative paths to specifications of other Components via relative paths, absolute paths, or URLs.

* **Crds** ([]string)

	Crds specifies relative paths to Custom Resource Definition files. This allows custom resources to be recognized as operands, making it possible to add them to the Resources list. CRDs themselves are not modified.

* **bases** ([]string)

	Deprecated: Anything that would have been specified here should be specified in the Resources field instead.

* **configMapGenerator** ([]ConfigMapArgs)

	ConfigMapGenerator is a list of configmaps to generate from local data (one configMap per list item). The resulting resource is a normal operand, subject to name prefixing, patching, etc.  By default, the name of the map will have a suffix hash generated from its contents.

* **secretGenerator** ([]SecretArgs)

	SecretGenerator is a list of secrets to generate from local data (one secret per list item). The resulting resource is a normal operand, subject to name prefixing, patching, etc.  By default, the name of the map will have a suffix hash generated from its contents.

* **helmGlobals** (HelmGlobals)

	HelmGlobals contains helm configuration that isn't chart specific.

* **helmCharts** ([]HelmChart)

	HelmCharts is a list of helm chart configuration instances.

* **helmChartInflationGenerator** ([]HelmChartArgs)

	Deprecated: Auto-converted to HelmGlobals and HelmCharts. HelmChartInflationGenerator is a list of helm chart configurations.

* **generatorOptions** (GeneratorOptions)

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

	BuildMetadata is a list of strings used to toggle different build options
