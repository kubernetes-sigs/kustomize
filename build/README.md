[releases page]: https://github.com/kubernetes-sigs/kustomize/releases
[`cloud-build-local`]: https://github.com/GoogleCloudPlatform/cloud-build-local
[Google Cloud Build]: https://cloud.google.com/cloud-build

Scripts and configuration files for publishing a
`kustomize` release on the [releases page].

### Build a release locally

Install [`cloud-build-local`], then run

```
cloud-build-local \
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

Define the version (see [semver principles](https://semver.org)), e.g.:
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

The new tag will trigger a job in [Google Cloud
Build] to put a new release on the [releases page].
