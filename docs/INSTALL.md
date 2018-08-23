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
go get github.com/kubernetes-sigs/kustomize
```

## Installation using the kustomize.sh wrapper

You can use the [kustomize.sh](scripts/kustomize.sh) script to
automatically download and cache the binary when kustomize is started.
Typically you could add that script in your project git repo, making it easier
to use kustomize for all developers and also from your ci/cd pipeline.

One advantage of this approach is that you can also better control what
version of kustomize is in use for your project.
