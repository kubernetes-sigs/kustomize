[release page]: https://github.com/kubernetes-sigs/kustomize/releases
[Go]: https://golang.org

# Installation

Binaries at various versions for linux, macOs and Windows
are available on the [release page].

Or...

## Quickly curl the latest binary

```
# pick one
opsys=darwin
opsys=windows
opsys=linux

curl -s https://api.github.com/repos/kubernetes-sigs/kustomize/releases/latest |\
  grep browser_download |\
  grep $opsys |\
  cut -d '"' -f 4 |\
  xargs curl -O -L
mv kustomize_kustomize\.v*_${opsys}_amd64 kustomize
chmod u+x kustomize
```

## Build the kustomize CLI from local source
```
# Need go 1.13 or higher
unset GOPATH
# see https://golang.org/doc/go1.13#modules
unset GO111MODULES

# clone the repo
git clone git@github.com:kubernetes-sigs/kustomize.git
# get into the repo root
cd kustomize

# Optionally checkout a particular tag if you don't
# want to build at head
git checkout kustomize/v3.2.3

# build the binary
(cd kustomize; go install .)

# run it
~/go/bin/kustomize version
```

### Other methods

#### Use go get

This works poorly with existing `Go` package installations at the
moment, since kustomize switched over to Go modules but hasn't
historically followed semver with respect to its API.

This is being [fixed](versioningPolicy.md), after which
`go get` should work correctly.

#### macOS

```
brew install kustomize
```

#### windows

```
choco install kustomize
```

For support on the chocolatey package
and prior releases, see:
- [Choco Package](https://chocolatey.org/packages/kustomize)
- [Package Source](https://github.com/kenmaglio/choco-kustomize)
