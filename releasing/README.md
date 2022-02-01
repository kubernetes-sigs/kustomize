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

#### Prepare your source directory

The release scripts expect Kustomize code to be cloned at a path ending in `sigs.k8s.io/kustomize`. Run all commands from that directory unless otherwise specified.

#### Consider fetching new OpenAPI data
The Kubernetes OpenAPI data changes no more frequently than once per quarter.
You can check the current builtin versions that kustomize is using with the
following command.

```
kustomize openapi info
```

Instructions on how to get a new OpenAPI sample can be found in the
[OpenAPI Readme].

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

While you're waiting for the tests, review the commit log. Based on the changes to be included in this release, decide whether a patch, minor or major version bump is needed: [semver review].

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
createBranch pinToKyaml "Pin to kyaml $versionKyaml"
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

While you're waiting for the tests, review the commit log. Based on the changes to be included in this release, decide whether a patch, minor or major version bump is needed: [semver review].

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
createBranch pinToCmdConfig "Pin to cmd/config $versionCmdConfig" &&
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

While you're waiting for the tests, review the commit log. Based on the changes to be included in this release, decide whether a patch, minor or major version bump is needed: [semver review].

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
createBranch pinToApi "Pin to api $versionApi" &&
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

While you're waiting for the tests, review the commit log. Based on the changes to be included in this release, decide whether a patch, minor or major version bump is needed: [semver review].

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


## Unpin everything


Go back into development mode, where all modules depend on in-repo code:

```
gorepomod unpin api         --doIt &&
gorepomod unpin cmd/config  --doIt &&
gorepomod unpin kyaml       --doIt
```

Create the PR:
```
createBranch unpinEverything "Back to development mode; unpin the modules" &&
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
new images. Assign the PR to @monopole and @Shell32-natsu.

## Update kustomize-in-kubectl

[kubernetes/kubernetes]: https://github.com/kubernetes/kubernetes
[newest kustomize releases]: https://github.com/kubernetes-sigs/kustomize/releases

To update the version of kustomize shipped with kubectl, first
fork and clone the [kubernetes/kubernetes] repo.

In the root of the kubernetes repo, run the following commands, modifying
the version numbers to match the [newest kustomize releases]:
```bash
./hack/pin-dependency.sh sigs.k8s.io/kustomize/kyaml v0.11.0
./hack/pin-dependency.sh sigs.k8s.io/kustomize/cmd/config v0.9.13
./hack/pin-dependency.sh sigs.k8s.io/kustomize/api v0.8.11
./hack/pin-dependency.sh sigs.k8s.io/kustomize/kustomize/v4 v4.2.0

./hack/update-vendor.sh
./hack/update-internal-modules.sh 
./hack/lint-dependencies.sh 
```

If needed, manually update the kustomize attachment points in the following files:

`staging/src/k8s.io/cli-runtime/pkg/resource/kustomizevisitor.go`

`staging/src/k8s.io/cli-runtime/pkg/resource/kustomizevisitor_test.go`

`staging/src/k8s.io/kubectl/pkg/cmd/kustomize/kustomize.go`

`staging/src/k8s.io/cli-runtime/pkg/resource/builder.go`

Here are some example PRs:

https://github.com/kubernetes/kubernetes/pull/103419

https://github.com/kubernetes/kubernetes/pull/106389

----
Older notes follow:

## Public Modules

[`sigs.k8s.io/kustomize/api`]: #sigsk8siokustomizeapi
[`sigs.k8s.io/kustomize/kustomize`]: #sigsk8siokustomizekustomize
[`sigs.k8s.io/kustomize/kyaml`]: #sigsk8siokustomizekyaml
[`sigs.k8s.io/kustomize/cmd/config`]: #sigsk8siokustomizecmdconfig

[kustomize/v3.2.1]: /../../releases/tag/kustomize%2Fv3.2.1

| Module Name                          | Module Description         | Example Tag         | Example Branch Name         |
| ------                               | ---                        | ---                 | ---                         |
| [`sigs.k8s.io/kustomize/kustomize`]  | kustomize executable       | _kustomize/v3.2.2_  | _release-kustomize-v3.2.2_  |
| [`sigs.k8s.io/kustomize/api`]        | kustomize API              | _api/v0.1.0_        | _release-api-v0.1_          |
| [`sigs.k8s.io/kustomize/kyaml`]      | k8s-specific yaml editting | _kyaml/v0.3.3_      | _release-kyaml-v0.2_        |
| [`sigs.k8s.io/kustomize/cmd/config`] | kyaml related commands     | _cmd/config/v0.3.3_ | _release-cmd/config-v0.3_   |

### sigs.k8s.io/kustomize/kustomize

The `kustomize` program, a CLI for performing
structured edits to Kubernetes Resource Model (KRM) YAML.

There's only one public package in this module.
It's called `main`, it's therefore unimportable.
It holds the `kustomize` executable.

See the [installation instructions](../docs/INSTALL.md)
to perform an install.


### sigs.k8s.io/kustomize/api

The [kustomize Go API](https://github.com/kubernetes-sigs/kustomize/tree/master/api).

The packages in this module are used by API clients,
like the `kustomize` program itself, and programs in
other repositories, e.g. `kubectl`.  They include
methods to read, edit and emit modified YAML.

Release notes should appear on the [release page].

The package has a toy executable called `api`, which,
if run, prints the API release version, but has no
other use.

### sigs.k8s.io/kustomize/kyaml

The [kyaml module](https://github.com/kubernetes-sigs/kustomize/tree/master/kyaml).

Low level packages for transforming YAML associated
with KRM objects.

These are used to build _kyaml filters_, computational units
that accept and emit KRM YAML, editing or simply validating it.

The kustomize `api` and the `cmd/config` packages are built on this.

### sigs.k8s.io/kustomize/cmd/config

The [cmd/config module](https://github.com/kubernetes-sigs/kustomize/tree/master/cmd/config).

A collection od CLI commands that correspond to
kyaml filters.

## Manual process

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

### Get up to date

It's assumed that whatever is already commited to the latest
commit is passing all tests and appropriate for release.


```
git fetch upstream
git checkout master
git rebase upstream/master
make prow-presubmit-check
```

### Select a module to release

E.g.
```
module=kustomize   # The kustomize executable
module=api         # The API
```

### Review tags to help determine new tag

Local:
```
git tag -l | grep $module
```

Remote:
```
git ls-remote --tags upstream | grep $module
```

Set the version you want:

```
major=0; minor=1; patch=0
```


### Create the release branch

Every module release occurs on it's own git branch.

The branch name doesn't include the patch number,
since the branch accumulates patch releases.

> TODO: define procedure for doing a cherrypick (committing a patch) to a
> release branch that already exists.

Name the branch:

```
branch="release-${module}-v${major}.${minor}"
echo "branch=$branch"
```

Create it:
```
git checkout -b $branch
```

### Define the release tag

```
tag="${module}/v${major}.${minor}.${patch}"
echo "tag=$tag"
```

### Pin modules to their dependencies.

This is achieved via a `replace` directive
in a module's `go.mod` file.

```
# Update the following as needed, obviously.

# git checkout -b pinTheRelease
# go mod edit -dropreplace=sigs.k8s.io/kustomize/api    $module/go.mod
# go mod edit -require=sigs.k8s.io/kustomize/api@v0.1.1 $module/go.mod
# git commit -a -m "Drop API module replacement"

```

### Push the release branch

```
git push -f upstream $branch
```

#### if replacing a release...

Must delete the tag before re-pushing it.
Dangerous - only do this if you're sure nothing
has already pulled the release.

Delete the tag locally:

```
git tag --delete $tag
```

Delete it upstream:
```
# Disable push protection:
git remote set-url --push upstream git@github.com:kubernetes-sigs/kustomize.git

# The empty space before the colon effectively means delete the tag.
git push upstream :refs/tags/$tag

# Enable push protection:
git remote set-url --push upstream no_push
```

Optionally visit the [release page] and delete
(what has now become) the _draft_ release for that
version.

### Tag the local repository

```
git tag -a $tag -m "Release $tag on branch $branch"
```

Move the `latest_kustomize` tag:
```
git tag -d latest_kustomize
git push upstream :latest_kustomize
git tag -a latest_kustomize
```

### Optionally build locally

[localbuild.sh]: localbuild.sh

Load the same version of `goreleaser` referenced in `cloudbuild.yaml` via docker and run [localbuild.sh] from the container's command line:

```
# Get goreleaser image from cloudbuild.yaml 
export GORELEASER_IMAGE=goreleaser/goreleaser:v0.172.1

# Drop into a shell
docker run -it --entrypoint=/bin/bash  -v $(pwd):/go/src/github.com/kubernetes-sigs/kustomize -w /go/src/github.com/kubernetes-sigs/kustomize $GORELEASER_IMAGE

# Run build
./releasing/localbuild.sh TAG [--snapshot]
```


### Optionally build and release locally

[cloudbuild-local.sh]: cloudbuild-local.sh

Install [`cloud-build-local`], then run [cloudbuild-local.sh]:

```
./releasing/cloudbuild-local.sh $module
```

This should create release artifacts in a local directory.

### Trigger the cloud build by pushing the tag

Push the tag:

```
git remote set-url --push upstream git@github.com:kubernetes-sigs/kustomize.git
git push upstream $tag
git push upstream latest_kustomize
git remote set-url --push upstream no_push
```

This triggers [Google Cloud Build].

###  Update release notes

Visit the [release page] and edit the release notes as desired.
