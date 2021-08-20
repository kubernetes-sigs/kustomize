#!/usr/bin/env bash

# If no argument is given -> Downloads the most recently released
# kustomize binary to your current working directory.
# (e.g. 'install_kustomize.sh')
#
# If one argument is given -> 
# If that argument is in the format of #.#.#, downloads the specified
# version of the kustomize binary to your current working directory.
# If that argument is something else, downloads the most recently released
# kustomize binary to the specified directory.
# (e.g. 'install_kustomize.sh 3.8.2' or 'install_kustomize.sh $(go env GOPATH)/bin')
#
# If two arguments are given -> Downloads the specified version of the
# kustomize binary to the specified directory.
# (e.g. 'install_kustomize.sh 3.8.2 $(go env GOPATH)/bin')
#
# Fails if the file already exists.

set -e

# Unset CDPATH to restore default cd behavior. An exported CDPATH can
# cause cd to output the current directory to STDOUT.
unset CDPATH

where=$PWD

function getTagName () {
    local res=""
    local modName="kustomize"
    
    # first version with kustomize binary in its own module
    local init="3.2.1"
    
    if [[ $1 == $init ]]
    then
        # res=kustomize/v$init
        echo "$modName/v$1"
        return
    fi
    
    local IFS=.
    local ver=($1) iVer=($init)
    
    for ((i=${#ver[@]}; i<${#iVer[@]}; i++))
    do
        ver[i]=0
    done
    
    for ((i=0; i<${#ver[@]}; i++))
    do
        if [[ -z ${iVer[i]} ]]
        then
            iVer[i]=0
        fi
        if ((10#${ver[i]} > 10#${iVer[i]}))
        then
            res=kustomize/v$1
            echo "$modName/v$1"
            return 0
        fi
        if ((10#${ver[i]} < 10#${iVer[i]}))
        then
            res=v$1
            echo "v$1"
            return 0
        fi
    done
}

release_url=https://api.github.com/repos/kubernetes-sigs/kustomize/releases
if [ -n "$1" ]; then
  if [[ "$1" =~ ^[0-9]+(\.[0-9]+){2}$ ]]; then
    tagName=$(getTagName $1)
    release_url=${release_url}/tags/$tagName
  elif [ -n "$2" ]; then
    echo "The first argument should be the requested version."
    exit 1
  else
    where="$1"
  fi
fi

if [ -n "$2" ]; then
  where="$2"
fi

if ! test -d "$where"; then
  echo "$where does not exist. Create it first."
  exit 1
fi

# Emulates `readlink -f` behavior, as this is not available by default on MacOS
# See: https://stackoverflow.com/questions/1055671/how-can-i-get-the-behavior-of-gnus-readlink-f-on-a-mac
function readlink_f {
  TARGET_FILE=$1

  cd "$(dirname "$TARGET_FILE")"
  TARGET_FILE=$(basename "$TARGET_FILE")

  # Iterate down a (possible) chain of symlinks
  while [ -L "$TARGET_FILE" ]
  do
      TARGET_FILE=$(readlink "$TARGET_FILE")
      cd "$(dirname "$TARGET_FILE")"
      TARGET_FILE=$(readlink "$TARGET_FILE")
  done

  # Compute the canonicalized name by finding the physical path
  # for the directory we're in and appending the target file.
  PHYS_DIR=$(pwd -P)
  RESULT=$PHYS_DIR/$TARGET_FILE
  echo "$RESULT"
}

where="$(readlink_f $where)/"

if [ -f "${where}kustomize" ]; then
  echo "${where}kustomize exists. Remove it first."
  exit 1
elif [ -d "${where}kustomize" ]; then
  echo "${where}kustomize exists and is a directory. Remove it first."
  exit 1
fi

tmpDir=`mktemp -d`
if [[ ! "$tmpDir" || ! -d "$tmpDir" ]]; then
  echo "Could not create temp dir."
  exit 1
fi

function cleanup {
  rm -rf "$tmpDir"
}

trap cleanup EXIT ERR

pushd "$tmpDir" >& /dev/null

opsys=windows
arch=amd64
if [[ "$OSTYPE" == linux* ]]; then
  opsys=linux
elif [[ "$OSTYPE" == darwin* ]]; then
  opsys=darwin
fi

RELEASE_URL=$(curl -s $release_url |\
  grep browser_download.*${opsys}_${arch} |\
  cut -d '"' -f 4 |\
  sort -V | tail -n 1)

if [ ! -n "$RELEASE_URL" ]; then
  echo "Version $version does not exist."
  exit 1
fi

curl -sLO $RELEASE_URL

tar xzf ./kustomize_v*_${opsys}_${arch}.tar.gz

cp ./kustomize "$where"

popd >& /dev/null

${where}kustomize version

echo "kustomize installed to ${where}kustomize"
