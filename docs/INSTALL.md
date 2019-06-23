[release page]: https://github.com/kubernetes-sigs/kustomize/releases
[Go]: https://golang.org

## Installation

For linux, macOs and Windows,
download a binary from the
[release page].

Or try this command:
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

To install from head with [Go] v1.12 or higher:

<!-- @installkustomize @test -->
```
go install sigs.k8s.io/kustomize/v3/cmd/kustomize
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
