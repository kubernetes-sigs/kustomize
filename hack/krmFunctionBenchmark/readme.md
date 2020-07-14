# Benchmark for KRM functions in Kustomize

## Prerequisites

 - [kustomize v3.7.0 or higher](https://kubernetes-sigs.github.io/kustomize/installation) (support for KRM Config Functions).
 - [docker](https://github.com/kubernetes-sigs/kustomize/pull/docker.com)

## How to run

```bash
./benchmark.sh
```

The script will build the exec version of function via container and then run 10, 100 and 1000 times of exec version and container version and then print out the time used by both versions.

```bash
./cleanup.sh
```
Will remove the built exec version of the function. Add flag `--image` to remove the images that used to build exec function.

## Functions in the benchmark

Two functions are used:

 - `gcr.io/kustomize-functions/example-tshirt` ([link](https://github.com/kubernetes-sigs/kustomize/blob/master/functions/examples/injection-tshirt-sizes/image/main.go))
 - `gcr.io/kpt-functions/label-namespace`	 ([link](https://github.com/GoogleContainerTools/kpt-functions-sdk/blob/master/ts/hello-world/src/label_namespace.ts))

 `example-tshirt` is a Go function. `label-namespace` is a JS function. Both of them are used as transformers.
