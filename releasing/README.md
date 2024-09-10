# Releasing

[release page]: /../../releases
[GitHub Actions]: /../../actions
[Google Cloud Build]: https://cloud.google.com/cloud-build
[semver]: https://semver.org
[Go modules]: https://github.com/golang/go/wiki/Modules
[multi-module repo]: https://github.com/go-modules-by-example/index/blob/master/009_submodules/README.md
[semver review]: #semver-review
[semver release]: #semver-review
[`cloudbuild_kustomize_image.yaml`]: cloudbuild_kustomize_image.yaml
[`release.yaml`]: ../.github/workflows/release.yaml
[`create-release.sh`]: create-release.sh
[kustomize repo release page]: https://github.com/kubernetes-sigs/kustomize/releases
[OpenAPI Readme]: ../kyaml/openapi/README.md
[the build status for container image]: https://console.cloud.google.com/cloud-build/builds?project=k8s-staging-kustomize

This document describes how to perform a [semver release]
of one of the several [Go modules] in this repository.

In this repo, releasing a module is accomplished by applying
a tag to the repo and pushing it upstream.  A minor release
branch is also created as necessary to track patch releases.

A properly formatted tag (described below) contains
the module name and version.

Pushing the tag upstream will trigger [GitHub Actions] to build a release and make it available on the  [release page].
[GitHub Actions] reads its instructions from the [`release.yaml`] file in `.github/workflows` directory.

And, container image contains `kustomize` binary will build [Google Cloud Build] that instructions from [`cloudbuild_kustomize_image.yaml`] file triggered by tags contain `kustomize` and release versions.

We use a Go program to make the tagging and branch
creation process less error prone.

See this [multi-module repo] tagging discussion
for an explanation of the path-like portion of these tags.

#### semver review

Go's [semver]-compatible version tags take the form `v{major}.{minor}.{patch}`:

| major | minor | patch |
| :---:   | :---:  | :---: |
| API change | enhancements | bug fixes |
| manual update | maybe auto-update | auto-update encouraged |

 - If there are only bug fixes or refactors, increment `patch` from whatever it is now.
 - If there are new features, increment  `minor`.
 - If there's an API change (either the Go API or the CLI behavior
   with respect to CLI arguments and flags), increment `major`.

## Release sequence

The dependencies determine the release order:

| module | depends on |
| ---    | ---        |
| `sigs.k8s.io/kustomize/kyaml`      | no local dependencies |
| `sigs.k8s.io/kustomize/cmd/config` | `kyaml`,              |
| `sigs.k8s.io/kustomize/api`        | `kyaml`               |
| `sigs.k8s.io/kustomize/kustomize`  | `cmd/config`, `api`   |

Thus, do `kyaml` first, then `cmd/config`, etc.

**⚠️ IMPORTANT:** The full release sequence should always be completed unless there is a really compelling reason not to AND the new release is fully backwards compatible with the latest released versions of our modules that depend on it (this is not detected by our CI, as we develop against the head versions of all modules). If we end up with a situation where the latest kyaml is not compatible with the latest api, for example, it causes end user confusion, particularly when automation attempts to upgrade kyaml for them.

## Prep work

### ⚠️ IMPORTANT: Check for [release-blocking issues](https://github.com/kubernetes-sigs/kustomize/issues?q=label%3Arelease-blocker+is%3Aopen)

We use the `release-blocker` tag to track issues that need to be solved before the next release. Typically, this would be a new regression introduced on the master branch and not present in the previous release. If any such issues exist, the release should be delayed.

It is also a good idea to scan any [untriaged issues](https://github.com/kubernetes-sigs/kustomize/issues?q=is%3Aissue+is%3Aopen+label%3Aneeds-triage) for potential blockers we haven't labelled yet before proceeding.

### Consider fetching new OpenAPI data

Ideally, Kustomize's embedded openapi data would cover a wide range of Kubernetes releases. But today, we only embed a specific version. This means updating that version can be disruptive to people who still use older Kubernetes versions and depend on API versions that were removed in later releases. However, by remaining out of date, we will not support GVKs introduced in more recent releases. So far, we have leaned in favour of the older versions, because some removed GVs are for very popular APIs. This should be constantly reevaluated until a better solution is in place. See issue https://github.com/kubernetes-sigs/kustomize/issues/5016.

The Kubernetes OpenAPI data changes no more frequently than once per quarter.
You can check the current builtin versions that kustomize is using with the
following command.

```
kustomize openapi info
```

Instructions on how to get a new OpenAPI sample can be found in the
[OpenAPI Readme].

### Set up the release tools

#### Prepare your source directory

The release scripts expect Kustomize code to be cloned at a path ending in `sigs.k8s.io/kustomize`. Run all commands from that directory unless otherwise specified.

#### Load some helper functions

```
source releasing/helpers.sh
```

#### Install the release tool

```
( cd cmd/gorepomod; go install . )
```

#### Authenticate to github using [gh](https://github.com/cli/cli) (version [1.8.1](https://github.com/cli/cli/releases/tag/v1.8.1) or higher).

```
gh auth login
```

## Releasing `kyaml`

Based on the changes to be included in this release, decide whether a patch, minor or major version bump is needed: [semver review].

To release `kyaml`, simply use command below.

```shell
./releasing/create-release kyaml [major|minor|patch]
```

The release command will go through these steps:
1. Create a release branch
2. Pull changes from master
3. Run Github Workflow for `kyaml`, which detailed into steps below.
   1. Run unit test
   2. Build library
   3. Bump `kyaml` version and push release tag
   4. Create a release draft
4. Maintainer will validate the release draft and publish

Undraft the release on the [kustomize repo release page]:
* Make sure the version number is what you expect.
* Remove references to commits that aren't relevant to end users of this module (e.g. test commits, refactors).
* Make sure each commit left in the release notes includes a PR reference.

## Releasing `cmd/config`

Based on the changes to be included in this release, decide whether a patch, minor or major version bump is needed: [semver review].

To release `cmd/config`, simply use command below.

```shell
./releasing/create-release cmd/config [major|minor|patch]
```

The release command will go through these steps:
1. Create a release branch in `release-cmd/config/vX.Y.Z` format
2. Pull changes from master
3. Run Github Workflow for `cmd/config`, which detailed into steps below.
   1. Run unit test
   2. Get latest tag from master branch and pin `kyaml` version
   3. Build library
   4. Bump `cmd/config` version and push release tag
   5. Create a release draft
4. Maintainer will validate the release draft and publish

Undraft the release on the [kustomize repo release page]:
* Make sure the version number is what you expect.
* Remove references to commits that aren't relevant to end users of this module (e.g. test commits, refactors).
* Make sure each commit left in the release notes includes a PR reference.

## Releasing `api`

Based on the changes to be included in this release, decide whether a patch, minor or major version bump is needed: [semver review].

To release `api`, simply use command below.

```shell
./releasing/create-release api [major|minor|patch]
```

The release command will go through these steps:
1. Create a release branch in `release-api/vX.Y.Z` format
2. Pull changes from master
3. Run Github Workflow for `api`, which detailed into steps below.
   1. Run unit test
   2. Get latest tag from master branch and pin `kyaml` version
   3. Build library
   4. Bump `api` version and push release tag
   5. Create a release draft
4. Maintainer will validate the release draft and publish

Undraft the release on the [kustomize repo release page]:
* Make sure the version number is what you expect.
* Remove references to commits that aren't relevant to end users of this module (e.g. test commits, refactors).
* Make sure each commit left in the release notes includes a PR reference.

## Releasing kustomize CLI

To release `kustomize`, simply use command below.

```shell
./releasing/create-release kustomize [major|minor|patch]
```

The release command will go through these steps:
1. Create a release branch in `release-kustomize/vX.Y.Z` format
2. Pull changes from master
3. Run Github Workflow for `kustomize`, which detailed into steps below.
   1. Run unit test
   2. Get latest tag from master branch and pin `cmd/config`, and `api` version
   3. Build library
   4. Bump `kustomize` version and push release tag
   5. Create a release draft
4. Maintainer will validate the release draft and publish

Undraft the release on the [kustomize repo release page]:
* Make sure the version number is what you expect.
* Remove references to commits that aren't relevant to end users of the CLI (e.g. test commits, refactors, changes that only surface in Go).
* Make sure each commit left in the release notes includes a PR reference.

#### For major releases: increment the module version

`Note:` This operation must be done manually to avoid error.

Update `module sigs.k8s.io/kustomize/kustomize/vX` in `kustomize/go.mod` to the version you're about to release, and then update all the `require` statements across the module to match.

Search for uses of the version number across the codebase and update them as needed.

Example: https://github.com/kubernetes-sigs/kustomize/pull/5021

## Confirm the kustomize binary is correct

[installation instructions]: https://kubectl.docs.kubernetes.io/installation/kustomize/binaries/

* Follow the [installation instructions] to install your new
  release and make sure it reports the expected version number.

  If not, something is very wrong.

* Visit the [release page] and edit the release notes as desired.

## Publish Official Docker Image

[k8s.io]: https://github.com/kubernetes/k8s.io
[k8s-staging-kustomize]: https://console.cloud.google.com/gcr/images/k8s-staging-kustomize?project=k8s-staging-kustomize

Fork and clone the [k8s.io] repo.

Checkout a new branch.

Edit file `registry.k8s.io/images/k8s-staging-kustomize/images.yaml`
to add the new kustomize version and the image sha256.

Image sha256 can be found in the image registry in the GCP project [k8s-staging-kustomize].

Commit and push your changes. Then create a PR to [k8s.io] to promote
the new image.

## Update kustomize-in-kubectl

[kubernetes/kubernetes]: https://github.com/kubernetes/kubernetes
[kustomize release]: https://github.com/kubernetes-sigs/kustomize/releases

To update the version of kustomize shipped with kubectl, first
fork and clone the [kubernetes/kubernetes] repo.

In the root of the [kubernetes/kubernetes] repo, run the following command:
```bash
./hack/update-kustomize.sh
```

When prompted, review the list of suggested module versions and make sure it matches the versions used in the latest [kustomize release]. If `kyaml`, `cmd/config` or `api` has been released more recently than `kustomize/v4`, **do not** use the more recent version.


If code used by the kustomize attachment points has changed, kubectl will fail to compile and you will need to update them. The code you'll need to change is likely in the `staging/src/k8s.io/cli-runtime/pkg/resource` and/or `staging/src/k8s.io/kubectl/pkg/cmd/kustomize` packages.

Here are some example PRs:

https://github.com/kubernetes/kubernetes/pull/103419

https://github.com/kubernetes/kubernetes/pull/106389


# Testing changes to the release pipeline

You can test the release script locally by running [`create-release.sh`].
See each of those files for more details on their usage.
