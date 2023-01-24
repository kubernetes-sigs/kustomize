#!/usr/bin/env bash
# Copyright 2023 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

read -a modules <<< $(go list -m)
read -a module_paths <<< $(go list -m -f {{.Dir}})

for i in ${!modules[@]}; do
  replace_path=$(realpath --relative-to=$(pwd) ${module_paths[i]})
  if [ $replace_path == . ]; then
    continue
  fi
  go mod edit -replace=${modules[i]}=$replace_path
done