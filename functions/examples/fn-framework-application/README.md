## Kyaml Functions Framework Example: Application Custom Resource

This is a moderate-complexity example of a KRM function built using the [KRM Functions Framework package](https://pkg.go.dev/sigs.k8s.io/kustomize/kyaml/fn/framework). It demonstrates how to write a function that implements a custom resource (CR) representing an abstract application. 

[embedmd]:# (pkg/exampleapp/v1alpha1/testdata/success/basic/config.yaml)
```yaml
apiVersion: platform.example.com/v1alpha1
kind: ExampleApp
metadata:
  name: simple-app-sample
env: production
workloads:
  webWorkers:
    - name: web-worker
      domains:
      - example.com
  jobWorkers:
    - name: job-worker
      replicas: 10
      resources: medium
      queues:
      - high
      - medium
      - low
    - name: job-worker-2
      replicas: 5
      queues:
      - bg2
datastores:
  postgresInstance: simple-app-sample-postgres
```

It also demonstrates the pattern of having the CR accept patches, allowing the user to customize the final result beyond the fields the CR exposes.

[embedmd]:# (pkg/exampleapp/v1alpha1/testdata/success/overrides/config.yaml)
```yaml
apiVersion: platform.example.com/v1alpha1
kind: ExampleApp
metadata:
  name: simple-app-sample
env: production
workloads:
  webWorkers:
    - name: web-worker
      domains:
      - first.example.com
    - name: web-worker-no-sidecar
      domains:
      - second.example.com

overrides:
  additionalResources:
    - custom-configmap.yaml
  resourcePatches:
    - web-worker-sidecar.yaml
  containerPatches:
    - custom-app-env.yaml
```

## Implementation walkthrough

The entrypoint for the function is [cmd/main.go](cmd/main.go), which invokes a ["dispatcher"](pkg/dispatcher/dispatcher.go) that determines which `Filter` implements the resource passed in. The dispatcher pattern allows a single function binary to handle multiple CRs, and is also useful for evolving a single CR over time (e.g. handle `ExampleApp` API versions `example.com/v1beta1` and `example.com/v1`).

[embedmd]:# (pkg/dispatcher/dispatcher.go go /.*VersionedAPIProcessor.*/ /}}/)
```go
	p := framework.VersionedAPIProcessor{FilterProvider: framework.GVKFilterMap{
		"ExampleApp": map[string]kio.Filter{
			"platform.example.com/v1alpha1": &v1alpha1.ExampleApp{},
		},
	}}
```


The ExampleApp type is defined in [pkg/exampleapp/v1alpha1/types.go](pkg/exampleapp/v1alpha1/types.go). It is responsible for implementing the logic of the CR, most of which is done by implementing the `kyaml.Filter` interface in [pkg/exampleapp/v1alpha1/processing.go](pkg/exampleapp/v1alpha1/processing.go). Internally, the filter function mostly builds up and executes a `framework.TemplateProcessor`. 

The ExampleApp type is annotated with [kubebuilder markers](https://book.kubebuilder.io/reference/markers/crd-validation.html), and a Go generator uses those to create the CRD YAML in [pkg/exampleapp/v1alpha1/platform.example.com_exampleapps.yaml](pkg/exampleapp/v1alpha1/platform.example.com_exampleapps.yaml). The CR then implements `framework.ValidationSchemaProvider`, which causes the CRD to be used for validation. It also implements `framework.Validator` to add custom validations and `framework.Defaulter` to add defaulting.

[embedmd]:# (pkg/exampleapp/v1alpha1/types.go go /.*type ExampleApp.*/ /}/)
```go
type ExampleApp struct {
	// Embedding these structs is required to use controller-gen to produce the CRD
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	// +kubebuilder:validation:Enum=production;staging;development
	Env string `json:"env" yaml:"env"`

	// +optional
	AppImage string `json:"appImage" yaml:"appImage"`

	Workloads Workloads `json:"workloads" yaml:"workloads"`

	// +optional
	Datastores Datastores `json:"datastores,omitempty" yaml:"datastores,omitempty"`

	// +optional
	Overrides Overrides `json:"overrides,omitempty" yaml:"overrides,omitempty"`
}
```

[embedmd]:# (pkg/exampleapp/v1alpha1/processing.go go /.*Filter.*/ /error\) {/)
```go
func (a ExampleApp) Filter(items []*yaml.RNode) ([]*yaml.RNode, error) {
```

[embedmd]:# (pkg/exampleapp/v1alpha1/processing.go go /.*Schema.*/ /error\) {/)
```go
func (a *ExampleApp) Schema() (*spec.Schema, error) {
```

[embedmd]:# (pkg/exampleapp/v1alpha1/processing.go go /.*Validate.*/ /error {/)
```go
func (a *ExampleApp) Validate() error {
```

[embedmd]:# (pkg/exampleapp/v1alpha1/processing.go go /.*Default.*/ /error {/)
```go
func (a *ExampleApp) Default() error {
```


## Running the Example

There are three ways to try this out:

A. Run `make example` in the root of the example to run the function with the test data in [pkg/exampleapp/v1alpha1/testdata/success/basic](pkg/exampleapp/v1alpha1/testdata/success/basic).

B. Run `go run cmd/main.go [FILE]` in the root of the example. Try it with the test input from one of the cases in [pkg/exampleapp/v1alpha1/testdata/success](pkg/exampleapp/v1alpha1/testdata/success). For example: `go run cmd/main.go pkg/exampleapp/v1alpha1/testdata/success/basic/config.yaml`.

C. Build the binary with `make build`, then run it with `app-fn [FILE]`. Try it with the test input from one of the cases in [pkg/exampleapp/v1alpha1/testdata/success](pkg/exampleapp/v1alpha1/testdata/success). For example: `app-fn pkg/exampleapp/v1alpha1/testdata/success/basic/config.yaml`.


