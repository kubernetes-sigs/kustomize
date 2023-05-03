#!/usr/bin/env bash
# Copyright 2023 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0


set -eo pipefail

# This generates a useful starting point for a CRD that can be used for validation.
echo "  - Generating CRD"
controller-gen crd paths=./... output:crd:dir=.

# controller-gen does not currently support "additionalProperties: false".
# This hack adds it manually to all properties sections of the schema.
echo "  - Adding additionalProperties: false to all properties sections"
if [[ "$OSTYPE" == linux* ]]; then
  # Linux (GNU sed) expression
  sed -i "s/\(^[[:space:]]*\)properties:/\1additionalProperties: false\n\1properties:/g" ./platform.example.com_exampleapps.yaml
elif [[ "$OSTYPE" == darwin* ]]; then
  # macOS (BSD sed) expression
  sed -i "" "s/\(^[[:space:]]*\)properties:/\1additionalProperties: false\n\1properties:/g" ./platform.example.com_exampleapps.yaml
else
  echo "Unsupported OS: $OSTYPE"
  exit 1
fi
