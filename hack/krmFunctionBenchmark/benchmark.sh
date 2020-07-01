#! /bin/bash

set -e

KUSTOMIZE_EXEC=kustomize
KUSTOMIZE_FLAGS="build --enable_alpha_plugins --enable-exec"

function build_label_namespace_exec {
    cd label_namespace/execfn
    . build.sh
    cd -
}

function build_tshirt_exec {
    cd example_tshirt/execfn
    . build.sh
    cd -
}

function build_exec {
    echo "Building exec functions..."
    build_tshirt_exec
    build_label_namespace_exec
    echo "Done. Start running benchmark."
}

function run_fn {
    local loop=$1
    local type=$2
    local dir=$3
    echo -e "=== Running ${type} ${loop} times ==="
    cd $dir
    local begin_time=$(date +%s%N)
    for ((i = 0; i < $loop; i++))
    do
        $KUSTOMIZE_EXEC $KUSTOMIZE_FLAGS > /dev/null
        echo -en "\r$i/$loop"
    done
    local end_time=$(date +%s%N)
    local time_diff=$(($end_time - $begin_time))
    local time_diff_s=$(echo "${time_diff} / 1000 / 1000 / 1000" | bc -l)
    echo -e "\n=== Time used: ==="
    echo "${time_diff_s}s"
    cd -
}

function run_label_namespace_benchmark {
    local loop=$1
    run_fn $loop "Label Namespace Transformer Exec Function" "label_namespace/execfn"
    run_fn $loop "Label Namespace Transformer Container Function" "label_namespace/containerfn"
}

function run_tshirt_benchmark {
    local loop=$1
    run_fn $loop "T-shirt Example Exec Function" "example_tshirt/execfn"
    run_fn $loop "T-shirt Example Container Function" "example_tshirt/containerfn"
}

loops=(10 100 1000)

build_exec

for l in "${loops[@]}"
do
    run_label_namespace_benchmark $l
    run_tshirt_benchmark $l
done