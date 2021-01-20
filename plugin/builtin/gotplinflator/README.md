

Built in inflator for go templates.

Working example (as exec go plugin) is here: https://github.com/epcim/gotplinflator

## TO TEST
  - rename SkipKinds to Kinds (and allow ! prefix)

## TODO

- better establish remoteResource struct
  - rename Dependencies to Deps
  - rename TemplatePatterns to TemplateGlob

- add valuesFile (ie: to share one file for all)

- fix go-getter to download repos

- maybe replace values with context, because "context->Values" in case of helmcharts

- SprigUtil.go
  - move all functions/logic here

- this file to be removed
