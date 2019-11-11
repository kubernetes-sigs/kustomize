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
brew install coreutils wget gnu-sed
```

Add the new tools to your PATH

## Run the pre-commit tests

Run the `pre-commit` target to verify your install

```sh
make pre-commit
```

This will run the go tests, as well as documentation tests.
