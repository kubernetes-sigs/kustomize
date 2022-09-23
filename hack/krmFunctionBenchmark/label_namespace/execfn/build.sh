#! /bin/bash
# Copyright 2022 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0


IMAGE_LABEL="label_namespace_build:latest"
BUILD_HOME=/usr/local/build

docker build -t $IMAGE_LABEL .

docker run --rm -v $(pwd):/out $IMAGE_LABEL \
    cp -r $BUILD_HOME/ts/hello-world/dist $BUILD_HOME/ts/hello-world/node_modules /out
