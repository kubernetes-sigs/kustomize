#!/bin/bash
#
# Works exactly like cloudbuild.sh but doesn't perform a release.
#
# Usage (from top of repo):
#
#  releasing/localbuild.sh TAG [--snapshot]
#
# Where TAG is in the form
#
#   api/v1.2.3
#   kustomize/v1.2.3
#   cmd/config/v1.2.3
#   ... etc.
#
# This script runs a build through goreleaser (http://goreleaser.com) but nothing else.
#

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

# Find previous tag that matches the tags module
prevTag=$(git tag -l "$module*" --sort=-version:refname --no-contains=$fullTag | head -n 1)

# Generate the changelog for this release
# using the last two tags for the module
changeLogFile=$(mktemp)
git log $prevTag..$fullTag \
  --pretty=oneline \
  --abbrev-commit --no-decorate --no-color --no-merges \
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

goReleaserConfigFile=$(mktemp)

cat <<EOF >$goReleaserConfigFile
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
  - arm64

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

cat $goReleaserConfigFile

date

time /usr/local/bin/goreleaser build \
  --debug \
  --timeout 10m \
  --parallelism 4 \
  --config=$goReleaserConfigFile \
  --rm-dist \
  --skip-validate $remainingArgs

date
