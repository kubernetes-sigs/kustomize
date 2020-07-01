# Benchmark for KRM functions in Kustomize

## Pre-request

 - You have to use Kustomize version with KRM function supported.
 - You need to have Docker running on your machine.

## How to run

```bash
./benchmark.sh
```

The script will build the exec version of function via container and then run 10, 100 and 1000 times of exec version and container version and then print out the time used by both versions.

```bash
./cleanup.sh
```
Will remove the built exec version of the function. Add flag `--image` to remove the images that used to build exec function.