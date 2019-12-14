#!/bin/bash
# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

set -e

: "${api_major?Need to source VERSIONS}"
: "${api_minor?Need to source VERSIONS}"
: "${api_patch?Need to source VERSIONS}"

: "${kyaml_major?Need to source VERSIONS}"
: "${kyaml_minor?Need to source VERSIONS}"
: "${kyaml_patch?Need to source VERSIONS}"

: "${cmd_config_major?Need to source VERSIONS}"
: "${cmd_config_minor?Need to source VERSIONS}"
: "${cmd_config_patch?Need to source VERSIONS}"

: "${cmd_kubectl_major?Need to source VERSIONS}"
: "${cmd_kubectl_minor?Need to source VERSIONS}"
: "${cmd_kubectl_patch?Need to source VERSIONS}"


# api
go mod edit -dropreplace=sigs.k8s.io/kustomize/api@v0.0.0
go mod edit -require=sigs.k8s.io/kustomize/api@v$api_major.$api_minor.$api_patch

# kyaml
go mod edit -dropreplace=sigs.k8s.io/kustomize/kyaml@v0.0.0
go mod edit -require=sigs.k8s.io/kustomize/kyaml@v$kyaml_major.$kyaml_minor.$kyaml_patch

# cmd/config
go mod edit -dropreplace=sigs.k8s.io/kustomize/cmd/config@v0.0.0
go mod edit -require=sigs.k8s.io/kustomize/cmd/config@v$cmd_config_major.$cmd_config_minor.$cmd_config_patch

# cmd/kubectl
go mod edit -dropreplace=sigs.k8s.io/kustomize/cmd/kubectl@v0.0.0
go mod edit -require=sigs.k8s.io/kustomize/cmd/kubectl@v$cmd_kubectl_major.$cmd_kubectl_minor.$cmd_kubectl_patch
