[release page]: https://github.com/kubernetes-sigs/kustomize/releases
[Go]: https://golang.org

# Installation

Binaries at various versions for linux, macOs and Windows
are available on the [release page].

Or...

## Quickly curl the latest

```
opsys=linux  # or darwin, or windows
curl -s https://api.github.com/repos/kubernetes-sigs/kustomize/releases/latest |\
  grep browser_download |\
  grep $opsys |\
  cut -d '"' -f 4 |\
  xargs curl -O -L
mv kustomize_*_${opsys}_amd64 kustomize
chmod u+x kustomize
```

## Install from the HEAD of master branch

Requires [Go] v1.12 or higher:

<!-- @installkustomize @testAgainstLatestRelease -->
```
go install sigs.k8s.io/kustomize/v3/cmd/kustomize
```

> With [Go v1.12](https://golang.org/doc/go1.12#modules), prefix the above command with `GO111MODULE=on`, e.g.
> ```
> GO111MODULE=on go install sigs.k8s.io/kustomize/v3/cmd/kustomize
> ```
> This shouldn't be necessary with [Go v1.13](https://golang.org/doc/go1.13#modules).

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
