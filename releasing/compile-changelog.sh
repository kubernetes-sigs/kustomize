#!/bin/bash
#
# Builds a PR-oriented changelog from the git history for the given module.
#
# Usage (from top of repo):
#
#  releasing/compile-changelog.sh MODULE TAG CHANGE_LOG_FILE
#
# Where TAG is in the form
#
#   api/v1.2.3
#   kustomize/v1.2.3
#   cmd/config/v1.2.3
#   ... etc.
#

set -o errexit
set -o nounset
set -o pipefail

if [[ -z "${1-}" ]] || [[ -z "${2-}" ]]; then
  echo "Usage: $0 <module> <fullTag> <changeLogFile>"
  echo "Example: $0 kyaml kyaml/v0.13.4 changelog.txt"
  exit 1
fi

module=$1
fullTag=$2
changeLogFile="${3:-}"

# Find previous tag that matches the tags module
allTags=$(git tag -l "$module*" --sort=-version:refname --no-contains="$fullTag")
prevTag=$(echo "$allTags" | head -n 1)
echo "Compiling $module changes from $prevTag to $fullTag"

commits=( $(git log "$prevTag".."$fullTag" \
  --pretty=format:'%H' \
  --abbrev-commit --no-decorate --no-color --no-merges \
  -- "$module") )

echo "Gathering PRs for commits: ${commits[*]}"

# There is a 256 character limit on the query parameter for the GitHub API, so split into batches then deduplicate results
batchSize=5
results=""
for((i=0; i < ${#commits[@]}; i+=batchSize))
do
  commitList=$(IFS="+"; echo "${commits[@]:i:batchSize}" | sed 's/ /+/g')
  if newResults=$(curl -sSL "https://api.github.com/search/issues?q=$commitList+repo%3Akubernetes-sigs%2Fkustomize" | jq -r '[  .items[] |  { number, title } ]'); then
    results=$(echo "$results" "$newResults" | jq -s '.[0] + .[1] | unique')
  else
    echo "Failed to fetch results for commits (exit code $?): $commitList"
    exit 1
  fi
done

changelog=$(echo "${results}" | jq -r '.[] | select( .title | startswith("Back to development mode") | not) | "#\(.number): \(.title)" ')

if [[ -n "$changeLogFile" ]]; then
  echo "$changelog" > "$changeLogFile"
else
  echo
  echo "----CHANGE LOG----"
  echo "$changelog"
fi
