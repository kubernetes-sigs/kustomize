[release page]: /../../releases
[Go]: https://golang.org
[script]: https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh
[quickly]: https://www.arp242.net/curl-to-sh.html

# Installation

Binaries at various versions for linux, macOs and Windows
are available for manual download on the [release page].

To do this [quickly], run, run:

```
curl -s "https://raw.githubusercontent.com/\
kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"  | bash
```

This [script]
 - attempts to detect your OS,
 - downloads and unpacks the release tar file in a temporary directory,
 - copies the kustomize binary (and only that file) to your current working directory,
 - removes the temporary directory.

## Try `go`

This method is more to show off how the `go` tool works,
than for any practical purpose.  A kustomize developer should
clone the repo (see next section), and CI/CD scripts should 
download a specific ready-to-run executable rather than
rely on the `go` tool.

Install the latest kustomize binary in the v3 series to `$GOPATH/bin`:
```
go install sigs.k8s.io/kustomize/kustomize/v3
```

Install a specific version:
```
go get sigs.k8s.io/kustomize/kustomize/v3@v3.3.0
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
