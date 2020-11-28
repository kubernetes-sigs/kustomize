#!/bin/bash
set -e

cd ../api/internal/target
go test -bench=.