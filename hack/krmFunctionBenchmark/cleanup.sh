#! /bin/bash
# Copyright 2022 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

set -e
echo "You may need to run as root to clean."

rm -rf example_tshirt/execfn/tshirt label_namespace/execfn/dist label_namespace/execfn/node_modules

if [ "$1" == "--image" ]; then
    docker image rm label_namespace_build:latest
    docker image rm tshirt_example_build:latest
fi

echo "Done"
