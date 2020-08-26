#!/bin/bash
#
# Usage (from top of repo):
#
#  releasing/cloudbuild.sh TAG [--snapshot]
#
# Where TAG is in the form
#
#   api/v1.2.3
#   kustomize/v1.2.3
#   cmd/config/v1.2.3
#   ... etc.
#
# Cloud build should be configured to trigger on tags
# matching:
#
#   [\w/]+/v\d+\.\d+\.\d+
#
# This script runs goreleaser (http://goreleaser.com),
# presumably from a cloudbuild.yaml step that installed it.

set -e
set -x

fullTag=$1
shift
echo "fullTag=$fullTag"

remainingArgs="$@"
echo "Remaining args:  $remainingArgs"

# Take everything before the last slash.
# This is expected to match $module.
module=${fullTag%/*}
echo "module=$module"

# Obtain most recent commit hash associated with the module.
lastCommitHash=$(
  git log --tags=$module -1 \
  --oneline --no-walk --pretty=format:%h)

# Generate the changelog for this release
# using commit hashes and commit messages.
changeLogFile=$(mktemp)
git log $lastCommitHash.. \
  --pretty=oneline \
  --abbrev-commit --no-decorate --no-color \
  -- $module > $changeLogFile
echo "Release notes:"
cat $changeLogFile

# Take everything after the last slash.
# This should be something like "v1.2.3".
semVer=`echo $fullTag | sed "s|$module/||"`
echo "semVer=$semVer"

# This is probably a directory called /workspace
echo "pwd = $PWD"

# Sanity check
echo "### ls -las . ################################"
ls -las .
# echo "### ls -C /usr/bin ################################"
# ls -C /usr/bin
echo "###################################"


# CD into the module directory.
# This directory expected to contain a main.go, so there's
# no need for extra details in the `build` stanza below.
cd $module

skipBuild=true
if [[ "$module" == "kustomize" || "$module" == "pluginator" ]]; then
  # If releasing a main program, don't skip the build.
  skipBuild=false
fi

configFile=$(mktemp)
cat <<EOF >$configFile
project_name: $module

archives:
- name_template: "${module}_${semVer}_{{ .Os }}_{{ .Arch }}"

builds:
- skip: $skipBuild

  ldflags: >
    -s
    -X sigs.k8s.io/kustomize/api/provenance.version={{.Version}}
    -X sigs.k8s.io/kustomize/api/provenance.gitCommit={{.Commit}}
    -X sigs.k8s.io/kustomize/api/provenance.buildDate={{.Date}}

  goos:
  - linux
  - darwin
  - windows

  goarch:
  - amd64

checksum:
  name_template: 'checksums.txt'

env:
- CGO_ENABLED=0
- GO111MODULE=on

release:
  github:
    owner: kubernetes-sigs
    name: kustomize
  draft: true

EOF

cat $configFile

/bin/goreleaser release \
  --config=$configFile \
  --release-notes=$changeLogFile \
  --rm-dist \
  --skip-validate $remainingArgs 
