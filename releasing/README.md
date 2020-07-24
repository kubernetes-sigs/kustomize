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

This document describes how to perform a [semver release]
of one of the [Go modules] in this repository.

Briefly:
- The repo HEAD is in a clean state (all tests passing).
- A specially formatted tag is applied to the repo and pushed upstream
  (on minor or major revisions there will also be a new branch).
- A tag trigger in [Google Cloud Build], reading instructions from
  [`cloudbuild.yaml`] in this directory, creates release artifacts
  as needed and makes them available on the [release page].

## Public Modules

TODO: _This list is incomplete._

[`sigs.k8s.io/kustomize/kustomize`]: #sigsk8siokustomizekustomize
[`sigs.k8s.io/kustomize/api`]: #sigsk8siokustomizeapi
[`sigs.k8s.io/kustomize/kustomize`]: #sigsk8siokustomizekustomize
[`sigs.k8s.io/kustomize/kyaml`]: #sigsk8siokustomizekyaml
[`sigs.k8s.io/kustomize/cmd/config`]: #sigsk8siokustomizecmdconfig

[kustomize/v3.2.1]: /../../releases/tag/kustomize%2Fv3.2.1
[pluginator/v1.0.0]: /../../releases/tag/pluginator%2Fv1.0.0

| Module Name                          | Module Description         | Example Tag         | Example Branch Name        |
| ------                               | ---                        | ---                 | ---                        |
| [`sigs.k8s.io/kustomize/kustomize`]  | kustomize executable       | _kustomize/v3.2.2_  | _release-kustomize-v3.2.2_ |
| [`sigs.k8s.io/kustomize/api`]        | kustomize API              | _api/v0.1.0_        | _release-api-v0.1_         |
| [`sigs.k8s.io/kustomize/kyaml`]      | k8s-specific yaml editting | _kyaml/v0.3.3_      | _release-kyaml-v0.2_       |
| [`sigs.k8s.io/kustomize/cmd/config`] | kyaml related commands     | _cmd/config/v0.3.3_ | _release-cmd/config-v0.3_  |

> ### sigs.k8s.io/kustomize/kustomize
>
> The `kustomize` program, a CLI for performing
> structured edits to Kubernetes Resource Model (KRM) YAML.
>
> There's only one public package in this module.
> It's called `main`, it's therefore unimportable.
> It holds the `kustomize` executable.
>
> See the [installation instructions](../docs/INSTALL.md)
> to perform an install.
>
>
> ### sigs.k8s.io/kustomize/api
>
> The [kustomize Go API](https://github.com/kubernetes-sigs/kustomize/tree/master/api).
>
> The packages in this module are used by API clients,
> like the `kustomize` program itself, and programs in
> other repositories, e.g. `kubectl`.  They include
> methods to read, edit and emit modified YAML.
>
> Release notes should appear on the [release page].
>
> The package has a toy executable called `api`, which,
> if run, prints the API release version, but has no
> other use.
>
> ### sigs.k8s.io/kustomize/kyaml
>
> The [kyaml module](https://github.com/kubernetes-sigs/kustomize/tree/master/kyaml).
>
> Low level packages for transforming YAML associated
> with KRM objects.
>
> These are used to build _kyaml filters_, computational units
> that accept and emit KRM YAML, editing or simply validating it.
>
> The kustomize `api` and the `cmd/config` packages are built on this.
>
> ### sigs.k8s.io/kustomize/cmd/config
>
> The [cmd/config module](https://github.com/kubernetes-sigs/kustomize/tree/master/cmd/config).
>
> A collection od CLI commands that correspond to
> kyaml filters.
>
> ### sigs.k8s.io/kustomize/pluginator
>
> The `pluginator` program, a code generator that
> converts Go plugins to conventional statically
> linkable library code.
>
> Only holds a `main`, and therefore unimportable.
> It holds the _pluginator_ executable.
>
> This binary is only of
> interest to someone writing a new builtin
> transformer or generator.  See the [plugin
> documentation](../docs/plugins).
> Its dependence on the API is  for
> plugin-related constants, not logic.

## Release a module

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

### List the latest tags


```bash
(cd releasing/releasing; go run . list)
```

Example output (this shows properly formatted tags):

```
api/v0.4.1
cmd/config/v0.3.0
cmd/resource/v0.0.2
cmd/kubectl/v0.1.0
kyaml/v0.3.0
kustomize/v3.6.1
pluginator/v2.1.0
```

See this [multi-module repo] tagging discussion
for an explanation of the path-like portion of these tags.

### Release

It's assumed that whatever is already commited to the latest
commit is passing all tests and appropriate for release.

Get up to date:

```
git fetch upstream
git checkout master
git rebase upstream/master
```

Then:

```
(cd releasing/releasing; go run . release {moduleName} {versionField})
```

Where

 * `{moduleName}` must be `kyaml`, `kustomize`, `cmd/config`, etc.
 * `{versionField}` must be `major`, `minor` or `patch` (see [semver review] below).

The specified version will be incremented, and
inferior version(s) will be set to zero.

E.g. if the current version is `v3.4.1`, incrementing `minor`
results in `v3.5.0`.

This program only applies a tag, and pushes it upstream.

To make this actually happen, add the `--no-dry-run` flag
to the `release` command.

###  update release notes

Visit the [release page] and edit the release notes as desired.


## Manual process

The release program does the following for you.

### select a module to release

```
module=kustomize   # The kustomize executable
module=api         # The API
```

### review tags to help determine new tag

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



### create the release branch

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

### define the release tag

```
tag="${module}/v${major}.${minor}.${patch}"
echo "tag=$tag"
```

### pin the executable to a particular API version

Only do this if releasing one of the
executables (kustomize or pluginator).

In this repository, an executable in development
on the master branch typically depends on the API
also in development on the master branch.  This is
achieved via a `replace` directive in the
executable's `go.mod` file.

A _released_ executable, however, must depend on a
specific release of the API.  For this reason,
it's typical, but not required, to release an
executable immediately after releasing the API,
updating the API version that the executable
requires.

```
# Update the following as needed, obviously.

# git checkout -b pinTheRelease
# go mod edit -dropreplace=sigs.k8s.io/kustomize/api    $module/go.mod
# go mod edit -require=sigs.k8s.io/kustomize/api@v0.1.1 $module/go.mod
# git commit -a -m "Drop API module replacement"

```

### push the release branch

```
git push -f upstream $branch
```

### if replacing a release...

Must delete the tag before re-pushing it.

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

### tag the local repository

```
git tag -a $tag -m "Release $tag on branch $branch"
```

Move the `latest_kustomize` tag:
```
git tag -d latest_kustomize
git push upstream :latest_kustomize
git tag -a latest_kustomize
```

### optionally build a release locally

[localbuild.sh]: localbuild.sh

Install [`cloud-build-local`], then run [localbuild.sh]:

```
./releasing/localbuild.sh $module
```

This should create release artifacts in a local directory.

### trigger the cloud build by pushing the tag

Push the tag:

```
git remote set-url --push upstream git@github.com:kubernetes-sigs/kustomize.git
git push upstream $tag
git push upstream latest_kustomize
git remote set-url --push upstream no_push
```

This triggers [Google Cloud Build].

###  update release notes

Visit the [release page] and edit the release notes as desired.
