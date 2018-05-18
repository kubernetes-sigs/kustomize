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
  kustomize build $target > $tmpDir/my.yaml
  [[ $? -eq 0 ]] || { exitWith "Failed to kustomize build"; }

  cat $tmpDir/my.yaml

  # Setting the namespace this way is a test-infra thing?
  kubectl config set-context \
    $(kubectl config current-context) --namespace=default

  kubectl apply -f $tmpDir/my.yaml
  [[ $? -eq 0 ]] || { exitWith "Failed to run kubectl apply"; }

  sleep 20
}

function tearDownCluster {
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
  local namespace=`getPodField '{.items[0].metadata.namespace}'`
  [[ -z $namespace ]] && { exitWith "Unable to get namespace"; }

  # Create a user ldif file
  local userFile="user.ldif"
  cat <<EOF >$tmpDir/$userFile
dn: cn=The Postmaster,dc=example,dc=org
objectClass: organizationalRole
cn: The Postmaster
EOF
  [[ -f $tmpDir/$userFile ]] || { exitWith "Failed to create $tmpDir/$userFile"; }

  kubectl cp $tmpDir/$userFile  $namespace/$podName:/tmp/$userFile || \
    { exitWith "Failed to copy ldif file to Pod $podName"; }

  rm $tmpDir/$userFile

  podExec \
    ldapadd \
    -x \
    -w admin \
    -H ldap://localhost \
    -D "cn=admin,dc=example,dc=org" \
    -f /tmp/$userFile
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

target=$1

echo Kustomizing: \"$target\"
ls $target

tmpDir=$(mktemp -d)

configureCluster

podName=`getPodField '{.items[0].metadata.name}'`
[[ -z $podName ]] && { exitWith "Unable to get pod name"; }

echo pod is $podName
containerName="ldap"

getUserCount; [[ $? -eq 0 ]] || { exitWith "Expected no users."; }

addUser || { exitWith "Failed to add a user"; }
getUserCount; [[ $? -eq 1 ]] || { exitWith "Couldn't find the new added user"; }

deleteAddedUser || { exitWith "Failed to delete the user"; }
getUserCount; [[ $? -eq 0 ]] || { exitWith "User has not been deleted."; }

tearDownCluster
