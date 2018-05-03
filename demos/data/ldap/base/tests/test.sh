#!/bin/bash

# This script validates that this package works as expected with kustomize.
# The validation makes sure following steps are correctly executed and the output is as expected
# - deploy a ldap server by the output of kustomize
# - add a user
# - query a user
# - delete a user
#
# This script should be called as
# test.sh <path to directory of Kube-manifest.yaml>
#
# Testing passes if exit code is 0
# Tesging fails if exit code is 1
set -x

function exit_with {
  local msg=$1
  echo >&2 ${msg}
  exit 1
}

# make sure kustomize and kubectl are available
command -v kustomize >/dev/null 2>&1 || { exit_with "Require kustomize but it's not installed.  Aborting."; }
command -v kubectl >/dev/null 2>&1 || { exit_with "Require kubectl but it's not installed.  Aborting."; }

# set namespace to default
kubectl config set-context $(kubectl config current-context) --namespace=default

# run kustomize
# kustomize build $1 | kubectl apply -f - || { exit_with "Failed to run kubectl apply"; }
echo Kustomizing \"$1\"
ls $1
kustomize build $1 > generatedResources.yaml
[[ $? -eq 0 ]] || { exit_with "Failed to kustomize build"; }
cat generatedResources.yaml
kubectl apply -f generatedResources.yaml
[[ $? -eq 0 ]] || { exit_with "Failed to run kubectl apply"; }
sleep 20

# get the pod and namespace
pod=$(kubectl get pods -l app=ldap  -o jsonpath='{.items[0].metadata.name}')
namespace=$(kubectl get pods -l app=ldap  -o jsonpath='{.items[0].metadata.namespace}')
container="ldap"
[[ -z ${pod} ]] && { exit_with "Pod is not started successfully"; }
[[ -z ${namespace} ]] && { exit_with "Couldn't get namespace for Pod ${pod}"; }

# create a user ldif file locally
ldiffile="user.ldif"
cat <<EOF >$ldiffile
dn: cn=The Postmaster,dc=example,dc=org
objectClass: organizationalRole
cn: The Postmaster
EOF
[[ -f ${ldiffile} ]] || { exit_with "Failed to create ldif file locally"; }

# add a user
pod_ldiffile="/tmp/user.ldif"
kubectl cp $ldiffile  ${namespace}/${pod}:${pod_ldiffile} || { exit_with "Failed to copy ldif file to Pod ${pod}"; }
kubectl exec ${pod} -c ${container} -- ldapadd -x -H ldap://localhost -D "cn=admin,dc=example,dc=org" -w admin \
-f ${pod_ldiffile} || { exit_with "Failed to add a user"; }

# query the added user
r=$(kubectl exec ${pod} -c ${container} -- ldapsearch -x -H ldap://localhost -b dc=example,dc=org \
-D "cn=admin,dc=example,dc=org" -w admin)
user_count=$(echo ${r} | grep "cn: The Postmaster" | wc -l)
[[ ${user_count} -eq 0 ]] && { exit_with "Couldn't find the new added user"; }

# delete the added user
kubectl exec ${pod} -c ${container} -- ldapdelete -v -x -H ldap://localhost "cn=The Postmaster,dc=example,dc=org" \
-D "cn=admin,dc=example,dc=org" -w admin || {  exit_with "Failed to delete the user"; }
r=$(kubectl exec ${pod} -c ${container} -- ldapsearch -x -H ldap://localhost -b dc=example,dc=org -D "cn=admin,dc=example,dc=org" -w admin)
user_count=$(echo ${r} | grep "cn: The Postmaster" | wc -l)
[[ ${user_count} -ne 0 ]] && { exit_with "The user hasn't been deleted."; }

# kubectl delete
kubectl delete -f generatedResources.yaml
rm $ldiffile