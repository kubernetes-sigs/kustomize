# Sampling New OpenAPI Data

[kyaml]: ../
[OpenAPI schema]: ./kubernetesapi/swagger.json
[home]: ../../

This document describes how to fetch OpenAPI data from
a particular kubernetes version number. 

### Fetching the Schema 
In the [kyaml] directory, fetch the schema
```
make schema
```

You can specify a specific version with the "API_VERSION" 
parameter. The default version is v1.19.1. Here is an
example for fetching the data for v1.14.1.
```
make schema API_VERSION=v1.14.1
```

This will update the [OpenAPI schema]. 

### Generating Swagger.go
In the [kyaml] directory, generate the swagger.go files.
```
make openapi
```

### Run all tests
In the [home] directory, run the tests.
```
make prow-presubmit-check >& /tmp/k.txt; echo $?
# The exit code should be zero; if not examine /tmp/k.txt
```