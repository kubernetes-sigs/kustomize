#!/usr/bin/env bash

binDir=bin
arch=amd64
baseExeName=kustomize
gitCommit=`git log -1 --format="%H"`
now=`date -u +'%Y-%m-%dT%H:%M:%SZ'`
ldFlags="-X sigs.k8s.io/kustomize/api/provenance.version=qlik-dev -X sigs.k8s.io/kustomize/api/provenance.gitCommit=${gitCommit} -X sigs.k8s.io/kustomize/api/provenance.buildDate=${now}"

rm -rf ${binDir}
for os in linux darwin windows; do
    mkdir -p ${binDir}/${os}

    exeName=${baseExeName}
    if [[ ${os} = "windows" ]]; then
        exeName=${baseExeName}.exe
    fi

    pushd kustomize
    GOOS=${os} GOARCH=${arch} go build -ldflags "${ldFlags}" -o ../${binDir}/${os}/${exeName} main.go
    popd

    pushd ${binDir}/${os}
    tar -czvf ../${baseExeName}_${os}_${arch}.tar.gz ${exeName}
    popd
done
