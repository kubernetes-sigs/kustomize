#!/bin/bash

set -e

if [[ "$PULL_NUMBER" -ne "" ]]; then
    cmd="$MYGOBIN/prchecker
    -owner=$REPO_OWNER
    -repo=$REPO_NAME
    -pr=$PULL_NUMBER
    $MODULES"


    echo $MYGOBIN
    echo $REPO_OWNER
    echo $REPO_NAME
    echo $PULL_NUMBER
    echo $MODULES
    eval $cmd
else
    echo "Multi-module check skipped. No PULL_NUMBER provided.

To run this check locally set PULL_NUMBER to the PR ID from GitHub."
fi