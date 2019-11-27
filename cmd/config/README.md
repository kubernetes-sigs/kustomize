# cmd/config

This package exists to expose config filters directly as cli commands for the purposes
of development of the kyaml package and as a reference implementation for using the libraries.

## Docs

All documentation is also built directly into the `config` command group using 
`kustomize help config`.

- [tutorials](docs/tutorials)
- [commands](docs/commands)
- [api-conventions](docs/api-conventions)

## Build Command

Build the `config` command under `GOBIN`

    $ make

## Run Tests

Run all tests, linting, etc and build

    $ make all