[release page]: https://github.com/kubernetes-sigs/kustomize/releases
[Go]: https://golang.org

## Installation

On macOS, you can install kustomize with Homebrew package
manager:

    brew install kustomize

For all operating systems, download a binary from the
[release page].

Or try this to grab the latest official release
using the command line:

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

To install from head with [Go] v1.10.1 or higher:

<!-- @installkustomize @test -->
```
go get sigs.k8s.io/kustomize
```
