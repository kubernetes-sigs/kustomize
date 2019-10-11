# Releasing

[release page]: https://github.com/kubernetes-sigs/kustomize/releases
[`cloud-build-local`]: https://github.com/GoogleCloudPlatform/cloud-build-local
[Google Cloud Build]: https://cloud.google.com/cloud-build
[semver]: https://semver.org
[Go modules]: https://github.com/golang/go/wiki/Modules

This document describes how to perform a [semver] release
of one of the [Go modules] in this repository.

These modules release independently.

## Module summaries

[`sigs.k8s.io/kustomize/kustomize`]: #sigsk8siokustomizekustomize
[`sigs.k8s.io/kustomize`]: #sigsk8siokustomize
[`sigs.k8s.io/kustomize/pluginator`]: #sigsk8siokustomizepluginator
[kustomize/v3.2.1]: https://github.com/kubernetes-sigs/kustomize/releases/tag/kustomize%2Fv3.2.1
[pluginator/v1.0.0]: https://github.com/kubernetes-sigs/kustomize/releases/tag/pluginator%2Fv1.0.0

| Module Description | Module Prefix | Tag | Branch Name | 
| ---                | ------        | --- | ---         |
| kustomize executable  | [`sigs.k8s.io/kustomize/kustomize`]  | _kustomize/v{major}.{minor}.{patch}_  | _release-kustomize-v{major}.{minor}_  |
| kustomize Go API      | [`sigs.k8s.io/kustomize`]            | _v{major}.{minor}.{patch}_            | _release-api-v{major}.{minor}_        |
| pluginator executable | [`sigs.k8s.io/kustomize/pluginator`] | _pluginator/v{major}.{minor}.{patch}_ | _release-pluginator-v{major}.{minor}_ |


### sigs.k8s.io/kustomize/kustomize

The `kustomize` program, a CLI for performing
k8s-aware structured edits to YAML in k8s resource
files.

#### Packages

There's only one package in this module.  It's called `main`,
and it holds the _kustomize_ executable.


#### Release artifacts

Executable files appear for various operating
systems on the [release page].  The tag
appears in the URL, e.g. [kustomize/v3.2.1].

See the [installation instructions](../docs/INSTALL.md)
to perform an install.


### sigs.k8s.io/kustomize

This is the kustomize Go API.

#### Packages

The packages in this module are used by API clients,
like the `kustomize` program itself, and programs
in other repositories, e.g. `kubectl`.

They include methods to read, edit and emit
modified YAML.

Go consumers of this API will have a `go.mod` file
_requiring_ this module at a particular tag, e.g.

```
require sigs.k8s.io/kustomize/v4 v4.0.1
```

#### Release artifacts

This is a Go library only release, so the only
artifact per se is the repo tag, in the form `v4.3.2`,
that API clients can `require` from their `go.mod` file.

Release notes should appear on the [release page].

There's an executable called `kustapiversion`, which, if
run, prints the API release provenance data, but it's of
no practical use to an API client.

### sigs.k8s.io/kustomize/pluginator

The `pluginator` program, a code generator that
converts Go plugins to conventional statically
linkable library code.

#### Packages

There's only one package in this module.  It's called `main`,
and it holds the _pluginator_ executable.

At the time of writing this binary is only of
interest to someone writing a new builtin
transformer or generator.  See the [plugin
documentation](../docs/plugins).

Its dependence on the API is primarily for
plugin-related constants, not logic, and will
only change if there's some change in how
plugins are constructed (presumably
infrequently).

#### Release artifacts

Executables appear on the [release page].
The tag appears in the URL, e.g. [pluginator/v1.0.0].

## Release procedure

At any given moment, the repository's master branch is
passing all its tests and contains code one could release.

### get up to date

```
git fetch upstream
git checkout master
git rebase upstream/master
```

### select a module to release

```
module="api" # The API
module="kustomize" # The kustomize executable
module="pluginator" # The pluginator executable
```

### determine the version

Go's [semver]-compatible version tags take the form `v{major}.{minor}.{patch}`:

| major | minor | patch |
| :---:   | :---:  | :---: |
| API change | enhancements | bug fixes |
| manual update | OK to auto-update | OK to auto-update |

 - If there are only bug fixes or refactors, increment `patch` from whatever it is now.
 - If there are new features, increment  `minor`.
 - If there's an API change (either the Go API or the CLI behavior
   with respect to CLI arguments and flags), increment `major`.

```
major=1
minor=2
patch=3
```

### create the release branch

Every module release occurs on it's own git branch.

The branch name doesn't include the patch number,
since the branch accumulates patch releases.

> TODO: define procedure for doing a cherrypick (committing a patch) to a
> release branch that already exists.

```
branch="release-${module}-v${major}.${minor}"
echo "branch=$branch"
git checkout -b $branch
```


### remove API replacements from go.mod

Only do this if releasing one of the executables.

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

if [ "$module" != "api" ]; then
  # go mod edit -dropreplace=sigs.k8s.io/kustomize/v3    $module/go.mod
  # go mod edit -require=sigs.k8s.io/kustomize/v4@v4.0.1 $module/go.mod
  # git commit -a -m "Drop API module replacement"
fi
```

### optionally build a release locally

Install [`cloud-build-local`], then run

```
./releasing/localbuild.sh (kustomize|pluginator|api)
```

This should create release artifacts in a local directory.

### push the release branch

```
git push -f upstream $branch
```

### optionally review tags

```
git tag -l
git ls-remote --tags upstream
```

### define the release tag

```
tag="v${major}.${minor}.${patch}"
if [ "$module" != "api" ]; then
  # must prefix the tag with submodule name
  tag="${module}/${tag}"
fi
echo "tag=$tag"
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

### trigger the cloud build by pushing the tag

Push the tag:

```
git remote set-url --push upstream git@github.com:kubernetes-sigs/kustomize.git
git push upstream $tag
git remote set-url --push upstream no_push
```

This triggers a job in [Google Cloud Build] to
put a new release on the [release page].

###  update release notes

Visit the [release page] and edit the release notes as desired.
