## Overview

This directory contains scripts and configuration files for publishing a
`kustomize` release on [release page](https://github.com/kubernetes-sigs/kustomize/releases)

## Steps to run build a release locally
Install container-builder-local from [github](https://github.com/GoogleCloudPlatform/container-builder-local). 

```sh
container-builder-local --config=build/cloudbuild_local.yaml --dryrun=false --write-workspace=/tmp/w .
```

You will find the build artifacts under `/tmp/w/dist` directory
