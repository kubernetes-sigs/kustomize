#! /bin/bash

IMAGE_LABEL="tshirt_example_build:latest"
BUILD_HOME=/usr/local/build

docker build -t $IMAGE_LABEL .

docker run --rm -v $(pwd):/out $IMAGE_LABEL \
    cp -r $BUILD_HOME/functions/examples/injection-tshirt-sizes/image/tshirt /out
