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

This document describes how to perform a [semver release]
of one of the several [Go modules] in this repository.

## Release sequence

The dependencies determine the release order:

| module | depends on |
| ---    | ---        |
| `sigs.k8s.io/kustomize/kyaml`      | no local dependencies |
| `sigs.k8s.io/cli-utils`            | `kyaml`               |
| `sigs.k8s.io/kustomize/cmd/config` | `kyaml`, `cli-utils`  |
| `sigs.k8s.io/kustomize/api`        | `kyaml`               |
| `sigs.k8s.io/kustomize/kustomize`  | `cmd/config`, `api`   |

Thus, do `kyaml` first, then `cli-utils`, etc.

#### Consider fetching new OpenAPI data
The Kubernetes OpenAPI data changes no more frequently than once per quarter.
You can check the current builtin versions that kustomize is using with the
following command.
```
kustomize openapi info
```

Instructions on how to get a new OpenAPI sample can be found in the
[OpenAPI Readme].

## Prep work

#### Make some helper functions

```
function createBranch {
  branch=$1
  echo "Making branch $branch : \"$title\""
  git branch -D $branch  # delete if it exists
  git checkout -b $branch
  git commit -a -m "$title"
  git push -f origin $branch
}
```

```
function createPr {
  gh pr create --title "$title" --body "ALLOW_MODULE_SPAN" --base master
}
```

```
function refreshMaster {
  git checkout master
  git fetch upstream
  git rebase upstream/master
}
```

```
function testKustomizeRepo {
  make prow-presubmit-check >& /tmp/k.txt
  local code=$?
  if [ $code -ne 0 ]; then
    echo "**** FAILURE ******************"
    tail /tmp/k.txt
  else
    echo "LGTM"
  fi
}
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
cd ~/gopath/src/sigs.k8s.io/kustomize
refreshMaster
testKustomizeRepo
```

kyaml has no intra-repo deps, so if the tests pass,
it can just be released.

Release it:

```
gorepomod release kyaml --doIt
```

Note the version:
```
versionKyaml=v0.10.6   # EDIT THIS!
```

Undraft the release on the [kustomize repo release page],
make sure the version number is what you expect.


## Release [`cli-utils`](https://github.com/kubernetes-sigs/cli-utils)

```
cd ../cli-utils
```

Pin to the new version of kyaml you just released:
```
go mod edit -require sigs.k8s.io/kustomize/kyaml@$versionKyaml
```

Test it
```
make test
make test-e2e
```

Create the PR
```
title="Pin cli-utils to kyaml $versionKyaml"
createBranch pinKyaml
createPr
```

Wait for tests to pass, then merge the PR:
```
gh pr status
gh pr merge -m
```

Release it:
```
git co master
git rebase upstream/master
gorepomod release {top} --doIt
```

Note the version:
```
versionCliUtils=v0.22.4     # EDIT THIS!
```

## Release `cmd/config`

```
cd ../kustomize
```

Pin to the most recent kyaml.

```
gorepomod pin kyaml --doIt
```

If there is a cli-utils dependence,
pin to the version of cli-utils you just created
([releases](https://github.com/kubernetes-sigs/cli-utils/releases))

```
(cd cmd/config; \
 go mod edit -require=sigs.k8s.io/cli-utils@$versionCliUtils)
```

Test it.

```
testKustomizeRepo
```

Create the PR:
```
title="Pin to kyaml $versionKyaml and cli-utils $versionCliUtils"
createBranch pinToKyamlAndCliUtils
createPr
```

Run local tests while GH runs tests in the cloud:
```
testKustomizeRepo
```

Wait for tests to pass, then merge the PR:
```
gh pr status
gh pr merge -m
```

Get back on master and do paranoia test:
```
refreshMaster
testKustomizeRepo
```

Release it:
```
gorepomod release cmd/config --doIt
```

Note the version:
```
versionCmdConfig=v0.8.8 # EDIT THIS!
```

Undraft the release on the [kustomize repo release page],
make sure the version number is what you expect.


## Release `api` 

This is the kustomize API, used by the kustomize CLI.


Pin to the new cmd/config:

```
gorepomod pin cmd/config --doIt
```

Create the PR:
```
title="Pin to cmd/config $versionCmdConfig"
createBranch pinToCmdConfig
createPr
```

Run local tests while GH runs tests in the cloud:
```
testKustomizeRepo
```

Wait for tests to pass, then merge the PR:
```
gh pr status  # rinse, repeat
gh pr merge -m
```

Get back on master and do paranoia test:
```
refreshMaster
testKustomizeRepo
```

Release it:
```
gorepomod release api --doIt
```

Note the version:
```
versionApi=v0.7.2   # EDIT THIS!
```

Undraft the release on the [kustomize repo release page],
make sure the version number is what you expect.


## Release the kustomize CLI

Pin to the new API:
```
gorepomod pin api --doIt
```

Create the PR:
```
title="Pin to api $versionApi"
createBranch pinToApi
createPr
```

Run local tests while GH runs tests in the cloud:
```
testKustomizeRepo
```

Wait for tests to pass, then merge the PR:
```
gh pr status  # rinse, repeat
gh pr merge -m
```

Get back on master and do paranoia test:
```
refreshMaster
testKustomizeRepo
```

Release it:
```
gorepomod release kustomize --doIt
```

Undraft the release on the [kustomize repo release page].

## Confirm the kustomize binary is correct

> [installation instructions]: https://kubectl.docs.kubernetes.io/installation/kustomize/binaries/
> 
>  * Follow the [installation instructions] to install your new
>    release and make sure it reports the expected version number.
>
>    If not, something is very wrong.
>
>  * Visit the [release page] and edit the release notes as desired.


## Unpin everything


Go back into development mode, where all modules depend on in-repo code:

```
gorepomod unpin api         --doIt
gorepomod unpin cmd/config  --doIt
gorepomod unpin kyaml       --doIt
```

Create the PR:
```
title="Back to development mode; unpin the modules"
createBranch unpinEverything
createPr
```

Run local tests while GH runs tests in the cloud:
```
testKustomizeRepo
```

Wait for tests to pass, then merge the PR:
```
gh pr status  # rinse, repeat
gh pr merge -m
```

Get back on master and do paranoia test:
```
refreshMaster
testKustomizeRepo
```

### Publish Official Docker Image

[k8s.io]: https://github.com/kubernetes/k8s.io
[k8s-staging-kustomize]: https://pantheon.corp.google.com/gcr/images/k8s-staging-kustomize?project=k8s-staging-kustomize

Fork and clone the [k8s.io] repo.

Checkout a new branch.

Edit file `k8s.gcr.io/images/k8s-staging-kustomize/images.yaml`
to add the new kustomize version and the image sha256.

Image sha256 can be found in the image registry in the GCP 
project [k8s-staging-kustomize].

Commit and push your changes. Then create a PR to [k8s.io] to promote
new images. Assign the PR to @monopole and @Shell32-natsu.

### Finally

[Makefile]: https://github.com/kubernetes-sigs/kustomize/blob/master/Makefile

Edit the `prow-presubmit-target` in the [Makefile]
to test examples against your new release.

----

----

Older notes follow:

## Public Modules

[`sigs.k8s.io/cli-utils`]: #sigsk8siocli-utils
[`sigs.k8s.io/kustomize/api`]: #sigsk8siokustomizeapi
[`sigs.k8s.io/kustomize/kustomize`]: #sigsk8siokustomizekustomize
[`sigs.k8s.io/kustomize/kyaml`]: #sigsk8siokustomizekyaml
[`sigs.k8s.io/kustomize/cmd/config`]: #sigsk8siokustomizecmdconfig

[kustomize/v3.2.1]: /../../releases/tag/kustomize%2Fv3.2.1

| Module Name                          | Module Description         | Example Tag         | Example Branch Name         |
| ------                               | ---                        | ---                 | ---                         |
| [`sigs.k8s.io/cli-utils`]            | General cli-utils          | _cli-utils/v0.20.2_ | _release-cli-utils-v0.20.2_ |
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

### Optionally build a release locally

[localbuild.sh]: localbuild.sh

Install [`cloud-build-local`], then run [localbuild.sh]:

```
./releasing/localbuild.sh $module
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
