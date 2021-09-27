# kyaml internal forks

## qri-io/starlib

This code is used by the starlark runtime. We copied it in to reduce the dependencies being brought over to kubectl by the kustomize integration. Should it need updating, do so via manual copy-paste.

## go-yaml/yaml

This code is used extensively by kyaml. It is a copy of upstream at a particular revision that kubectl is using, with fixes we need cherry-picked on top ([#753](https://github.com/go-yaml/yaml/pull/753)). For background information on this problem, see https://github.com/kubernetes-sigs/kustomize/issues/3946.

This copy was created using the [git subtree technique](https://medium.com/@porteneuve/mastering-git-subtrees-943d29a798ec) and can be recreated on top of a new version of go-yaml v3 using the [update-go-yaml.sh](update-go-yaml.sh) script. To add an additional go-yaml PR to be cherry-picked, simply update the script's `GO_YAML_PRS` variable. Please note that there is nothing special about the fork directory, so copy-paste with manual edits will work just fine if you prefer.
