#!/usr/bin/env bash
# Copyright 2023 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

set -x
set -e
set -o pipefail
set -o nounset

# This script uses 'replace' statements to 'unpin' local modules from module versions go.mod normally
# requires, so that the local version will be used instead. With the advent of Workspace mode, we no longer
# need to do this in general in between releases. However, some key commands like `go mod tidy` are not
# Workspace-aware and thus will fail if API changes between modules exist on master. This script allows us to
# test those commands without requiring unpin operations in our release workflow.

if [[ -z "${1-}" ]] ; then
  echo "Usage: $0 <cmd>"
  echo "Example: $0 'go mod tidy -v'"
  exit 1
fi

cmd=$1

# First we read in the list of all kustomize modules and their local locations. The data looks like:
# sigs.k8s.io/kustomize/api /Users/you/src/sigs.k8s.io/kustomize/api
# sigs.k8s.io/kustomize/cmd/config /Users/you/src/sigs.k8s.io/kustomize/cmd/config
IFS=$'\n'
modules=($(go list -m -f "{{.Path}} {{.Dir}}"))

# Next we iterate over the lines, split apart the module name and local absolute path,
# and add a relative-path replace statement to the go.mod. A replace statement will be added
# for each Kustomize module, whether or not the current module uses it.
IFS=" "
replace_args=""
for module in "${modules[@]}"; do
  read -a module_data <<< $module
  replace_path=$(realpath --relative-to=$(pwd) ${module_data[1]})
  if [ $replace_path == . ] || [[ $replace_path == internal/* ]]; then
    continue
  fi
  replace_args+=" -replace=${module_data[0]}=$replace_path"
done

go mod edit $replace_args

# Now that the modules are pinned, we run the command passed to this script.
bash -c "$cmd"

# Finally we clean up by dropping the replace statements we added above.
go mod edit $(sed 's/-replace/-dropreplace/g' <<< "$replace_args" | sed -E 's/=\.\.[^[:space:]]*//g')
