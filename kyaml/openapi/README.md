# Sampling New OpenAPI Data

[OpenAPI schema]: ./kubernetesapi/swagger.json

This document describes how to fetch OpenAPI data from
a particular kubernetes version number.

### Fetching the Schema

In this directory, fetch the openapi schema for the kubernetes api:

```
make nuke
make kubernetesapi/swagger.json
```

You can specify a specific version with the "API_VERSION"
parameter. The default version is v1.19.1. Here is an
example for fetching the data for v1.14.1.

```
make kubernetesapi/swagger.json API_VERSION=v1.14.1
```

This will update the [OpenAPI schema].

### Generating Swagger.go

In this directory, generate the swagger.go files.

```
make
```

### Run all tests

At the top of the repository, run the tests.

```
make prow-presubmit-check >& /tmp/k.txt; echo $?
# The exit code should be zero; if not examine /tmp/k.txt
```
