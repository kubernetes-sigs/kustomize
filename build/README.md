## Steps to run build locally

Install container-builder-local from github and define GOOS, GOARCH, OUTPUT env
variables and run following command

```sh
container-builder-local --config=cmd/kustomize/build/cloudbuild_local.yaml --dryrun=false --substitutions=_GOOS=$GOOS,_GOARCH=$GOARCH --write-workspace=$OUTPUT .
```

## Steps to submit build to Google container builder

You need to be at the repo level to be able to run the following command

```
gcloud container builds submit . --config=cmd/kustomize/build/cloudbuild.yaml --substitutions=_GOOS=$GOOS,_GOARCH=$GOARCH
```
