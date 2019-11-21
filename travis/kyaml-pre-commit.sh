#!/bin/bash
set -e

cd kyaml
make all

cd ../cmd/config
make all