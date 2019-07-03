# Releasing

[release page]: https://github.com/kubernetes-sigs/kustomize/releases
[`cloud-build-local`]: https://github.com/GoogleCloudPlatform/cloud-build-local
[Google Cloud Build]: https://cloud.google.com/cloud-build
[semver]: https://semver.org

Scripts and configuration files for publishing a
`kustomize` release on the [release page].

## Build a release locally

Install [`cloud-build-local`], then run

```
./releasing/localbuild.sh
```

to build artifacts under `./dist`.

## Do a real (cloud) release

Get on an up-to-date master branch:
```
git fetch upstream
git checkout master
git rebase upstream/master
```

### review tags

```
git tag -l
git ls-remote --tags upstream
```

### define the new tag

Define the version per [semver] principles; it must start with `v`:

```
version=v3.0.0-pre
```

### if replacing a release...

Must delete the tag before re-pushing it.

Delete the tag locally:

```
git tag --delete $version
```

Delete it upstream:
```
# Disable push protection:
git remote set-url --push upstream git@github.com:kubernetes-sigs/kustomize.git

# The empty space before the colon effectively means delete the tag.
git push upstream :refs/tags/$version

# Enable push protection:
git remote set-url --push upstream no_push
```

Optionally visit the [release page] and delete
(what has now become) the _draft_ release for that
version.

### tag locally

```
git tag -a $version -m "Release $version"
```

### trigger the cloud build
Push the tag:
```
git push upstream $version
```

This triggers a job in [Google Cloud Build] to
put a new release on the [release page].

###  Update release notes

Visit the [release page] and edit the release notes as desired.

