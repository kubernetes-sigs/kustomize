# submodule

This repo demonstrates kustomize's ability to download git repos 
with submodules. The following branches contain
* main: submodule via absolute path
* relative-submodule: submodule via relative path

For the submodule accessed via a relative path, we include a random hash in the
submodule name to avoid accessing an unintended directory in the case kustomize
contains loader bugs (issue #5131).