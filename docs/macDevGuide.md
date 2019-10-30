# Developing on Mac OS

## Setup your dev environment

First install the tools to build and run tests

### Install go 1.13

[Instructions](https://golang.org/doc/install)

Add `go` to your PATH

### Install kubeval

[Instructions](https://github.com/instrumenta/kubeval)

```sh
go get github.com/instrumenta/kubeval
```

Add `kubeval` to your PATH

### Install gnu tools

[Instructions](https://www.topbug.net/blog/2013/04/14/install-and-use-gnu-command-line-tools-in-mac-os-x/)

```sh
brew install coreutils wget sed
```

Add the new tools to your PATH

## Clone Kustomize

- Kustomize must be cloned under `$GOPATH/src/sigs.k8s.io/` or some of the tests will not work

## Running the tests

Run the `pre-commit.sh` script to verify your install

```sh
./travis/pre-commit.sh
```

This will run the go tests, as well as documentation tests.

## Known Issues

`pre-commit.sh` will modify all api/plugins by changing `linux` to `darwin`.
See [#1711](https://github.com/kubernetes-sigs/kustomize/issues/1711) for details.

Don't check these updates in.