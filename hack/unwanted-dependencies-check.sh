#!/usr/bin/env bash
# Copyright 2024 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

set -o pipefail
set -o nounset
set +u

declare -i rc=0
declare -a POSTIONAL_ARGS=()

declare -x GO11MODULES=yes
declare -x GOFLAGS=-mod=mod

# Explicit path of the unwanted depedency list
JSON_PATH_URL=""
JSON_PATH_LOCAL=""
READ_PATH=""

parse_args() {
    while [[ "$1" != "" ]]; do
        case $1 in
            -u | --url )
                shift
                # Use json outside of repository
                JSON_PATH_URL="${1}"
                ;;
            -f | --file )
                shift
                # Use local json file
                JSON_PATH_LOCAL="${1}"
                ;;
            *)
                ;;
        esac
        shift
    done
}

check_requirements() {

    if [[ -z $(which jq) ]]; then
        rc=1
        echo "Error: jq not found, please install jq."
        exit ${rc}
    fi

    if [[ -z $(which wget) ]]; then
        rc=1
        echo "Error: wget not found, please install wget."
        exit ${rc}
    fi
}

pull_unwanted_dependencies_json() {

    if [[ ! -z "${JSON_PATH_URL}" ]]; then
        echo "${JSON_PATH_URL}"
        # Expected to be executed from root
        wget "${JSON_PATH_URL}" -O ${PWD}/hack/unwanted-dependencies.json 
        READ_PATH=${PWD}/hack/unwanted-dependencies.json
    elif [[ ! -z "${JSON_PATH_LOCAL}" ]]; then
        echo "${JSON_PATH_LOCAL}"
        # Expected to be executed from root
        JSON_PATH_LOCAL=$(realpath "${JSON_PATH_LOCAL}")
        if [[ -z $(stat ) ]]; then
            rc=1
            echo "Error: block list not supplied, please define block list file path."
            exit ${rc}
        fi
        READ_PATH=$(realpath ${JSON_PATH_LOCAL})
    else
        # Default behavior: pull unwanted-dependencies.json from kubernetes/kubernetes upstream repo
        JSON_PATH_URL='https://raw.githubusercontent.com/kubernetes/kubernetes/e51fe4a61cca7f4a0875630da433f280b52c138a/hack/unwanted-dependencies.json'
        wget "${JSON_PATH_URL}" -O "${PWD}/hack/unwanted-dependencies.json"
        READ_PATH="${PWD}/hack/unwanted-dependencies.json"
    fi
}

check_unwanted_dependencies(){
    for dep in $(jq -r '.spec.unwantedModules | keys[]' "${READ_PATH}"); do
        for file in $(find -f ./* | fgrep go.sum); do
            if [[ -z $(cat $file | fgrep $dep) ]]; then
                rc=1
                echo "Error: unwanted dependencies found. ($dep at $(realpath $file))"
            fi
        done
    done

    for upstream in $(jq -r '.status.unwantedReferences | keys[]' "${READ_PATH}"); do
        for ref in $(jq -r '.status.unwantedReferences.'\"${upstream}\"'[]' "${READ_PATH}"); do
            if [[ ! -z $(go mod graph | fgrep $upstream | fgrep $ref) ]]; then
                rc=1
                echo "Error: unwanted references found on one of the dependencies. ($upstream depends on $ref))"
            fi
        done
    done

    exit $rc
}

main() {
    parse_args $@
    check_requirements
    pull_unwanted_dependencies_json
    check_unwanted_dependencies
}

main $@