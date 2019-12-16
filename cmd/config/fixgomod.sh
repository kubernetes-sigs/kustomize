#!/bin/bash
# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

set -e

: "${kyaml_major?Need to source VERSIONS}"
: "${kyaml_minor?Need to source VERSIONS}"
: "${kyaml_patch?Need to source VERSIONS}"

go mod edit -dropreplace=sigs.k8s.io/kustomize/kyaml@v0.0.0
go mod edit -require=sigs.k8s.io/kustomize/kyaml@v$kyaml_major.$kyaml_minor.$kyaml_patch
