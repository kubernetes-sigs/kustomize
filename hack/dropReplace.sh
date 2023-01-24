#!/usr/bin/env bash
# Copyright 2023 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

read -a modules <<< $(go list -m)

for i in ${!modules[@]}; do
  go mod edit -dropreplace=${modules[i]}
done