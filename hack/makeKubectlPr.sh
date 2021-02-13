repoDir=$HOME/gopath/src/k8s.io/kubernetes
k8sio=staging/src/k8s.io

function saveIt {
  mkdir -p ~/safe/$1
  cp $1/$2 ~/safe/$1
}

function getIt {
  cp ~/safe/$1/$2 $1/$2
}

function doSave {
    rm -rf ~/safe
    mkdir ~/safe

    saveIt  ${k8sio}/cli-runtime/pkg/resource  builder.go
    saveIt  ${k8sio}/cli-runtime/pkg/resource  kustomizevisitor.go
    saveIt  ${k8sio}/cli-runtime/pkg/resource  kustomizevisitor_test.go
    saveIt  ${k8sio}/cli-runtime/pkg/resource  visitor.go

    saveIt  ${k8sio}/kubectl/pkg/cmd/kustomize kustomize.go

    echo "--- Saved: ----"
    tree ~/safe
}

function doReset {
    cd $repoDir
    git reset 11a05eb9aff104c2384781c740e053907a3849e6 --hard
    git clean -fd
    git log -n 1
    git status
}

function doCommit1 {
    git mv \
	${k8sio}/cli-runtime/pkg/kustomize/builder_test.go \
        ${k8sio}/cli-runtime/pkg/resource/kustomizevisitor_test.go

    git rm -r ${k8sio}/cli-runtime/pkg/kustomize
    git rm    ${k8sio}/kubectl/pkg/cmd/kustomize/kustomize_test.go

    git add --all
    git commit -m "Delete ${k8sio}/cli-runtime/pkg/kustomize"
}

function doCommit2 {
    getIt   ${k8sio}/cli-runtime/pkg/resource  builder.go
    getIt   ${k8sio}/cli-runtime/pkg/resource  kustomizevisitor.go
    getIt   ${k8sio}/cli-runtime/pkg/resource  kustomizevisitor_test.go
    getIt   ${k8sio}/cli-runtime/pkg/resource  visitor.go

    getIt   ${k8sio}/kubectl/pkg/cmd/kustomize kustomize.go

    (cd ${k8sio}/kubectl;     go mod tidy)
    (cd ${k8sio}/kubectl;     go test ./...)

    (cd ${k8sio}/cli-runtime; go mod tidy)
    (cd ${k8sio}/cli-runtime; go test ./...)

    (cd ${k8sio}/kubectl;     go mod tidy)
    (cd ${k8sio}/cli-runtime; go mod tidy)

    (cd ${k8sio}/kubectl;     go mod tidy)

    go mod edit --dropreplace=sigs.k8s.io/kustomize
    # go mod edit --exclude=sigs.k8s.io/kustomize@v2.0.3+incompatible

    git add --all
    git commit -m "Manually update kustomize attachment points."
}

function doCommit3 {
    ./hack/update-vendor.sh
    git add --all
    git commit -m "Run ./hack/update-vendor.sh"
}

function makePrBranch {
    doReset
    doCommit1
    doCommit2
    doCommit3
}
