## Overview

This directory contains scripts and configuration files for publishing a
`kustomize` release on [release page](https://github.com/kubernetes-sigs/kustomize/releases)

## Steps to run build a release locally
Install container-builder-local from [github](https://github.com/GoogleCloudPlatform/container-builder-local). 

```sh
container-builder-local --config=build/cloudbuild_local.yaml --dryrun=false --write-workspace=/tmp/w .
```

You will find the build artifacts under `/tmp/w/dist` directory

### Publishing a Release

Here are the steps to publish a new release.

 - Figure out the next version (for ex. v1.0.3) to be released.
 - Ensure that you are on master branch and is up-to-date with the remote upstream/master branch. Do the following if you are in doubt:
   ```
	git checkout master
	git fetch upstream
	git rebase upstream/master
   ```
 - Tag the repo by running `git tag -a <version> -m "<version> release"`
 - Push the tag to `upstream` branch by running `git push upstream <version>`
 - Thats it! The new tag will trigger a job in Google Container Builder and new release should appear in the releases page.
