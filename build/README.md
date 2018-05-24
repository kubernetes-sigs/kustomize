[releases page]: https://github.com/kubernetes-sigs/kustomize/releases
[`container-builder-local`]: https://github.com/GoogleCloudPlatform/container-builder-local
[Google Container Builder]: https://cloud.google.com/container-builder

Scripts and configuration files for publishing a
`kustomize` release on the [releases page].

### Build a release locally

Install [`container-builder-local`], then run

```
container-builder-local \
   --config=build/cloudbuild_local.yaml \
   --dryrun=false --write-workspace=/tmp/w .
```

to build artifacts under `/tmp/w/dist`.

### Publish a Release

Get on an up-to-date master branch:
```
git checkout master
git fetch upstream
git rebase upstream/master
```

Decide on your version, e.g.:
```
version=v1.0.3
```

Tag the repo:
```
git tag -a $version -m "$version release"
```

Push the tag upstream:
```
git push upstream $version
```

The new tag will trigger a job in [Google Container
Builder] to put a new release on the [releases page].
