#!/bin/bash
set -e

cd kyaml
make all

cd ../cmd/cfg
make all