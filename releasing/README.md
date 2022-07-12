# Releasing

[release page]: /../../releases
[`cloud-build-local`]: https://github.com/GoogleCloudPlatform/cloud-build-local
[Google Cloud Build]: https://cloud.google.com/cloud-build
[semver]: https://semver.org
[Go modules]: https://github.com/golang/go/wiki/Modules
[multi-module repo]: https://github.com/go-modules-by-example/index/blob/master/009_submodules/README.md
[semver review]: #semver-review
[semver release]: #semver-review
[`cloudbuild.yaml`]: cloudbuild.yaml
[kustomize repo release page]: https://github.com/kubernetes-sigs/kustomize/releases
[OpenAPI Readme]: ../kyaml/openapi/README.md
[project cloud build history page]: https://console.cloud.google.com/cloud-build/builds?project=k8s-staging-kustomize

This document describes how to perform a [semver release]
of one of the several [Go modules] in this repository.

In this repo, releasing a module is accomplished by applying
a tag to the repo and pushing it upstream.  A minor release
branch is also created as necessary to track patch releases.

A properly formatted tag (described below) contains
the module name and version.

Pushing the tag upstream will trigger [Google Cloud Build] to build a release
and make it available on the  [release page].

Cloud build reads its instructions from the
[`cloudbuild.yaml`] file in this directory.

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

## Prep work

### ⚠️ IMPORTANT: Check for [release-blocking issues](https://github.com/kubernetes-sigs/kustomize/issues?q=label%3Arelease-blocker+is%3Aopen)

We use the `release-blocker` tag to track issues that need to be solved before the next release. Typically, this would be a new regression introduced on the master branch and not present in the previous release. If any such issues exist, the release should be delayed.

It is also a good idea to scan any [untriaged issues](https://github.com/kubernetes-sigs/kustomize/issues?q=is%3Aissue+is%3Aopen+label%3Aneeds-triage) for potential blockers we haven't labelled yet before proceeding.

### Consider fetching new OpenAPI data
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
# Use your own token
GITHUB_TOKEN=deadbeefdeadbeef

echo $GITHUB_TOKEN | gh auth login --scopes repo --with-token
```

## Release `kyaml`

#### Establish clean state

```
refreshMaster &&
testKustomizeRepo
```

While you're waiting for the tests, review the commit log:

```
releasing/compile-changelog.sh kyaml HEAD 
```

Based on the changes to be included in this release, decide whether a patch, minor or major version bump is needed: [semver review].

kyaml has no intra-repo deps, so if the tests pass,
it can just be released.

#### Release it

The default increment is a new patch version.

```
gorepomod release kyaml [patch|minor|major] --doIt
```

Note the version:
```
versionKyaml=v0.10.20   # EDIT THIS!
```

See the process of the cloud build job
on the [project cloud build history page].

Undraft the release on the [kustomize repo release page]:
* Make sure the version number is what you expect.
* Remove references to commits that aren't relevant to end users of this module (e.g. test commits, refactors).
* Make sure each commit left in the release notes includes a PR reference.


## Release `cmd/config`

#### Pin to the most recent kyaml

```
gorepomod pin kyaml --doIt &&
go mod edit -require=sigs.k8s.io/kustomize/kyaml@$versionKyaml plugin/builtin/prefixtransformer/go.mod &&
go mod edit -require=sigs.k8s.io/kustomize/kyaml@$versionKyaml plugin/builtin/suffixtransformer/go.mod &&
go mod edit -require=sigs.k8s.io/kustomize/kyaml@$versionKyaml plugin/builtin/replicacounttransformer/go.mod &&
go mod edit -require=sigs.k8s.io/kustomize/kyaml@$versionKyaml plugin/builtin/patchtransformer/go.mod &&
go mod edit -require=sigs.k8s.io/kustomize/kyaml@$versionKyaml plugin/builtin/patchjson6902transformer/go.mod
```

Create the PR:
```
createBranch pinToKyaml "Update kyaml to $versionKyaml"
```
```
createPr
```

Run local tests while GH runs tests in the cloud:
```
testKustomizeRepo
```

Wait for tests to pass, then merge the PR:
```
gh pr status
```
```
gh pr merge -m
```

Get back on master and do paranoia test:
```
refreshMaster &&
testKustomizeRepo
```

While you're waiting for the tests, review the commit log:

```
releasing/compile-changelog.sh cmd/config HEAD 
```

Based on the changes to be included in this release, decide whether a patch, minor or major version bump is needed: [semver review].

#### Release it

```
gorepomod release cmd/config [patch|minor|major] --doIt
```

Note the version:
```
versionCmdConfig=v0.9.12 # EDIT THIS!
```

See the process of the cloud build job
on the [project cloud build history page].

Undraft the release on the [kustomize repo release page]:
* Make sure the version number is what you expect.
* Remove references to commits that aren't relevant to end users of this module (e.g. test commits, refactors).
* Make sure each commit left in the release notes includes a PR reference.


## Release `api`

This is the kustomize API, used by the kustomize CLI.


#### Pin to the new cmd/config

```
gorepomod pin cmd/config --doIt
```

Create the PR:
```
createBranch pinToCmdConfig "Update cmd/config to $versionCmdConfig" &&
createPr
```

Run local tests while GH runs tests in the cloud:
```
testKustomizeRepo
```

Wait for tests to pass, then merge the PR:
```
gh pr status  # rinse, repeat
```
```
gh pr merge -m
```

Get back on master and do paranoia test:
```
refreshMaster &&
testKustomizeRepo
```

While you're waiting for the tests, review the commit log:

```
releasing/compile-changelog.sh api HEAD 
```

Based on the changes to be included in this release, decide whether a patch, minor or major version bump is needed: [semver review].

#### Release it

```
gorepomod release api [patch|minor|major] --doIt
```

Note the version:
```
versionApi=v0.8.10 # EDIT THIS!
```

See the process of the cloud build job
on the [project cloud build history page].

Undraft the release on the [kustomize repo release page]:
* Make sure the version number is what you expect.
* Remove references to commits that aren't relevant to end users of this module (e.g. test commits, refactors).
* Make sure each commit left in the release notes includes a PR reference.


## Release the kustomize CLI

#### Pin to the new API

```
gorepomod pin api --doIt
```

Create the PR:
```
createBranch pinToApi "Update api to $versionApi" &&
createPr
```

Run local tests while GH runs tests in the cloud:
```
testKustomizeRepo
```

Wait for tests to pass, then merge the PR:
```
gh pr status  # rinse, repeat
```
```
gh pr merge -m
```

Get back on master and do paranoia test:
```
refreshMaster &&
testKustomizeRepo
```

While you're waiting for the tests, review the commit log:

```
releasing/compile-changelog.sh kustomize HEAD 
```

Based on the changes to be included in this release, decide whether a patch, minor or major version bump is needed: [semver review].

#### Release it

```
gorepomod release kustomize [patch|minor|major] --doIt
```

See the process of the cloud build job
on the [project cloud build history page].

Undraft the release on the [kustomize repo release page]:
* Make sure the version number is what you expect.
* Remove references to commits that aren't relevant to end users of the CLI (e.g. test commits, refactors, changes that only surface in Go).
* Make sure each commit left in the release notes includes a PR reference.

## Confirm the kustomize binary is correct

[installation instructions]: https://kubectl.docs.kubernetes.io/installation/kustomize/binaries/

* Follow the [installation instructions] to install your new
  release and make sure it reports the expected version number.

  If not, something is very wrong.

* Visit the [release page] and edit the release notes as desired.

## Update example test target

[Makefile]: https://github.com/kubernetes-sigs/kustomize/blob/master/Makefile

Edit the `prow-presubmit-target` in the [Makefile]
to test examples against your new release.

```
sed -i "" "s/LATEST_V4_RELEASE=.*/LATEST_V4_RELEASE=v4.3.0/" Makefile
```
```
createBranch updateProwExamplesTarget "Test examples against latest release" &&
createPr
```

Wait for tests to pass, then merge the PR:
```
gh pr status  # rinse, repeat
```
```
gh pr merge -m
```


## Publish Official Docker Image

[k8s.io]: https://github.com/kubernetes/k8s.io
[k8s-staging-kustomize]: https://console.cloud.google.com/gcr/images/k8s-staging-kustomize?project=k8s-staging-kustomize

Fork and clone the [k8s.io] repo.

Checkout a new branch.

Edit file `k8s.gcr.io/images/k8s-staging-kustomize/images.yaml`
to add the new kustomize version and the image sha256.

Image sha256 can be found in the image registry in the GCP 
project [k8s-staging-kustomize].

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

You can test the release script locally by running [cloudbuild.sh](cloudbuild.sh) in a container or by installing Cloud Build Local and running [cloudbuild-local.sh](cloudbuild-local.sh). See each of those files for more details on their usage.
