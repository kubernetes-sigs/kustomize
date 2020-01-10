#!/bin/bash
# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

set -e

: "${kyaml_major?Need to source VERSIONS}"
: "${kyaml_minor?Need to source VERSIONS}"
: "${kyaml_patch?Need to source VERSIONS}"

: "${kstatus_major?Need to source VERSIONS}"
: "${kstatus_minor?Need to source VERSIONS}"
: "${kstatus_patch?Need to source VERSIONS}"


go mod edit -dropreplace=sigs.k8s.io/kustomize/kyaml@v0.0.0
go mod edit -require=sigs.k8s.io/kustomize/kyaml@v$kyaml_major.$kyaml_minor.$kyaml_patch

go mod edit -dropreplace=sigs.k8s.io/kustomize/kstatus@v0.0.0
go mod edit -require=sigs.k8s.io/kustomize/kstatus@v$kstatus_major.$kstatus_minor.$kstatus_patch
