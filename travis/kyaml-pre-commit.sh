#!/bin/bash
set -e

cd kyaml
make all

cd ../cmd/kyaml
make all