#!/usr/bin/env bash
# Copyright 2024 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0


declare PATCH=false
declare MINOR=false
declare MAJOR=false
declare rc=0

ORIGIN_MASTER="origin/master"
LATEST_TAG=$(git describe --tags --abbrev=0)

git log "${LATEST_TAG}..HEAD" --oneline | tee /tmp/release-changelogs.txt

count=$(cat /tmp/release-changelogs.txt | wc -l)

if [[ $(cat /tmp/release-changelogs.txt | grep fix) || $(cat /tmp/release-changelogs.txt | grep patch) || $(cat /tmp/release-changelogs.txt | grep chore) || $(cat /tmp/release-changelogs.txt | grep docs) ]]; then
    PATCH=true
fi

if [[ $(cat /tmp/release-changelogs.txt | grep feat) ]]; then
    MINOR=true
fi

for commit in $(cut -d' ' -f1 /tmp/release-changelogs.txt); do
    git log --format=%B -n 1 $commit | grep "BREAKING CHANGE"
    if [ $? -eq 0 ]; then
        MAJOR=true
    fi
done

for f in $(find api); do
    git diff "${LATEST_TAG}...${ORIGIN_MASTER}" --exit-code -- "${f}" 
    if [ $? -eq 1 ]; then
        echo "Found changes on api dir at ${f}"
        MAJOR=true
    fi
done

echo -e "\n"
echo -e "================================================================================="
echo "Change counter: $(echo $count | tr -s ' ')"
if [[ $MAJOR == false && $MINOR == false ]]; then
    echo "Recommended release type: patch"
elif [[ $MAJOR == false && $MINOR == true ]]; then
    echo "Recommended release type: minor"
else
    echo "Recommended release type: major"
fi