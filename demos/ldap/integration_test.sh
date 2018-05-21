#!/bin/bash
# Copyright 2018 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# ----------------------------------------------------
#
# This script tests the ldap kustomization demo
# against a real cluster.
#
# - deploy a ldap server by the output of kustomize
# - add a user
# - query a user
# - delete a user
#
# This script is a test that (passes|fails) if exit
# code is (0|1).

set -x

function configureCluster {
  local tmpDir=$1
  local target=$2

  echo Kustomizing: \"$target\"
  ls $target

  kustomize build $target > $tmpDir/my.yaml
  [[ $? -eq 0 ]] || { exitWith "Failed to kustomize build"; }

  # Setting the namespace this way is a test-infra thing?
  kubectl config set-context \
    $(kubectl config current-context) --namespace=default

  kubectl apply -f $tmpDir/my.yaml
  [[ $? -eq 0 ]] || { exitWith "Failed to run kubectl apply"; }

  sleep 20
}

function tearDownCluster {
  local tmpDir=$1
  kubectl delete -f $tmpDir/my.yaml
  rm -rf $tmpDir
}

function getPodField {
  echo $(kubectl get pods -l app=ldap -o jsonpath=$1)
}

function podExec {
  kubectl exec $podName -c $containerName -- "$@"
}

function addUser {
  local tmpDir=$1
  local namespace=`getPodField '{.items[0].metadata.namespace}'`
  [[ -z $namespace ]] && { exitWith "Unable to get namespace"; }

  local myUserFile=$tmpDir/user.ldif
  local podUserFile=/tmp/user.ldif

  cat <<EOF >$myUserFile
dn: cn=The Postmaster,dc=example,dc=org
objectClass: organizationalRole
cn: The Postmaster
EOF
  [[ -f $myUserFile ]] || \
    { exitWith "Failed to create $myUserFile"; }

  kubectl cp $myUserFile \
    $namespace/$podName:$podUserFile || \
    { exitWith "Failed to copy $myUserFile to pod $podName"; }

  rm $myUserFile

  podExec \
    ldapadd \
    -x \
    -w admin \
    -H ldap://localhost \
    -D "cn=admin,dc=example,dc=org" \
    -f $podUserFile
}

function getUserCount {
  local result=$(\
    podExec \
      ldapsearch \
      -x \
      -w admin \
      -H ldap://localhost \
      -D "cn=admin,dc=example,dc=org" \
      -b dc=example,dc=org \
      )
  return $(echo $result | grep "cn: The Postmaster" | wc -l)
}

function deleteAddedUser {
  podExec \
    ldapdelete \
    -v \
    -x \
    -w admin \
    -H ldap://localhost \
    -D "cn=admin,dc=example,dc=org" \
    "cn=The Postmaster,dc=example,dc=org"
}

tmpDir=$(mktemp -d)

configureCluster $tmpDir $1

podName=`getPodField '{.items[0].metadata.name}'`
[[ -z $podName ]] && { exitWith "Unable to get pod name"; }
containerName="ldap"

getUserCount; [[ $? -eq 0 ]] || { exitWith "Expected no users."; }

addUser $tmpDir || { exitWith "Failed to add a user"; }
getUserCount; [[ $? -eq 1 ]] || { exitWith "Couldn't find the new added user"; }

deleteAddedUser || { exitWith "Failed to delete the user"; }
getUserCount; [[ $? -eq 0 ]] || { exitWith "User has not been deleted."; }

tearDownCluster $tmpDir
