#!/usr/bin/env bash

declare PATCH=false
declare MINOR=false
declare MAJOR=false
declare rc=0

git log $(git describe --tags --abbrev=0)..HEAD --oneline | tee /tmp/release-changelogs.txt

if [[ $(cat /tmp/release-changelogs.txt | grep fix) || $(cat /tmp/release-changelogs.txt | grep patch) ]]; then
    PATCH=true
fi

if [[ $(cat /tmp/release-changelogs.txt | grep feat) ]]; then
    MINOR=true
fi

for f in $(find api); do
    git diff --exit-code "${f}"
    if [ $? -eq 1 ]; then
        echo "Found changes on api dir at ${f}"
        rc=1
        exit 1
    fi
done

if [ $rc -eq 1 ]; then
    MAJOR=true
fi

echo -e "\n"
echo -e "================================================================================="

if [[ $MAJOR == false && $MINOR == false ]]; then
    echo "Release type: patch"
elif [[ $MAJOR == false && $MINOR == true ]]; then
    echo "Release type: minor"
else
    echo "Release type: major"
fi