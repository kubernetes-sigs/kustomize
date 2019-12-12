#!/bin/bash
set -e

cd kyaml
make all

cd ../cmd/config
make all

cd ../kubectl
make all

# make sure no files were generated or changed by make
cd ../..
git add .
git diff-index HEAD --
