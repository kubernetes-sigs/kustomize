# Copyright 2022 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

### Kustomize plugin rules.
#
# The rules to deal with builtin plugins are a bit
# complicated because
#
# - Every builtin plugin is a Go plugin -
#   meaning it gets its own module directory
#   (outside of the api module) with Go
#   code in a 'main' package per Go plugin rules.
# - kustomize locates plugins using the
#   'apiVersion' and 'kind' fields from the
#   plugin config file.
# - k8s wants CamelCase in 'kind' fields.
# - The module name (the last name in the path)
#   must be the lowercased 'kind' of the
#   plugin because Go and related tools
#   demand lowercase in import paths, but
#   allow CamelCase in file names.
# - the generated code must live in the api
#   module (it's linked into the api).

# Where all generated builtin plugin code should go.
pGen=api/internal/builtins
# Where the builtin Go plugin modules live.
pSrc=plugin/builtin

_builtinplugins = \
	AnnotationsTransformer.go \
	ConfigMapGenerator.go \
	IAMPolicyGenerator.go \
	HashTransformer.go \
	ImageTagTransformer.go \
	LabelTransformer.go \
	LegacyOrderTransformer.go \
	NamespaceTransformer.go \
	PatchJson6902Transformer.go \
	PatchStrategicMergeTransformer.go \
	PatchTransformer.go \
	PrefixTransformer.go \
	SuffixTransformer.go \
	ReplacementTransformer.go \
	ReplicaCountTransformer.go \
	SecretGenerator.go \
	ValueAddTransformer.go \
	HelmChartInflationGenerator.go

# Maintaining this explicit list of generated files, and
# adding it as a dependency to a few targets, to assure
# they get recreated if deleted.  The rules below on how
# to make them don't, by themselves, assure they will be
# recreated if deleted.
builtinplugins = $(patsubst %,$(pGen)/%,$(_builtinplugins))

# These rules are verbose, but assure that if a source file
# is modified, the corresponding generated file, and only
# that file, will be recreated.
$(pGen)/AnnotationsTransformer.go: $(pSrc)/annotationstransformer/AnnotationsTransformer.go
$(pGen)/ConfigMapGenerator.go: $(pSrc)/configmapgenerator/ConfigMapGenerator.go
$(pGen)/GkeSaGenerator.go: $(pSrc)/gkesagenerator/GkeSaGenerator.go
$(pGen)/HashTransformer.go: $(pSrc)/hashtransformer/HashTransformer.go
$(pGen)/ImageTagTransformer.go: $(pSrc)/imagetagtransformer/ImageTagTransformer.go
$(pGen)/LabelTransformer.go: $(pSrc)/labeltransformer/LabelTransformer.go
$(pGen)/LegacyOrderTransformer.go: $(pSrc)/legacyordertransformer/LegacyOrderTransformer.go
$(pGen)/NamespaceTransformer.go: $(pSrc)/namespacetransformer/NamespaceTransformer.go
$(pGen)/PatchJson6902Transformer.go: $(pSrc)/patchjson6902transformer/PatchJson6902Transformer.go
$(pGen)/PatchStrategicMergeTransformer.go: $(pSrc)/patchstrategicmergetransformer/PatchStrategicMergeTransformer.go
$(pGen)/PatchTransformer.go: $(pSrc)/patchtransformer/PatchTransformer.go
$(pGen)/PrefixTransformer.go: $(pSrc)/prefixtransformer/PrefixTransformer.go
$(pGen)/SuffixTransformer.go: $(pSrc)/suffixtransformer/SuffixTransformer.go
$(pGen)/ReplacementTransformer.go: $(pSrc)/replacementtransformer/ReplacementTransformer.go
$(pGen)/ReplicaCountTransformer.go: $(pSrc)/replicacounttransformer/ReplicaCountTransformer.go
$(pGen)/SecretGenerator.go: $(pSrc)/secretgenerator/SecretGenerator.go
$(pGen)/ValueAddTransformer.go: $(pSrc)/valueaddtransformer/ValueAddTransformer.go
$(pGen)/HelmChartInflationGenerator.go: $(pSrc)/helmchartinflationgenerator/HelmChartInflationGenerator.go

# The (verbose but portable) Makefile way to convert to lowercase.
toLowerCase = $(subst A,a,$(subst B,b,$(subst C,c,$(subst D,d,$(subst E,e,$(subst F,f,$(subst G,g,$(subst H,h,$(subst I,i,$(subst J,j,$(subst K,k,$(subst L,l,$(subst M,m,$(subst N,n,$(subst O,o,$(subst P,p,$(subst Q,q,$(subst R,r,$(subst S,s,$(subst T,t,$(subst U,u,$(subst V,v,$(subst W,w,$(subst X,x,$(subst Y,y,$(subst Z,z,$1))))))))))))))))))))))))))

$(pGen)/%.go: $(MYGOBIN)/pluginator
	@echo "generating $*"
	( \
		set -e; \
		cd $(pSrc)/$(call toLowerCase,$*); \
		go generate .; \
		cd ../../../$(pGen); \
		$(MYGOBIN)/goimports -w $*.go \
	)

# Target is for debugging.
.PHONY: generate-kustomize-builtin-plugins
generate-kustomize-builtin-plugins: $(builtinplugins)

.PHONY: build-kustomize-external-go-plugin
build-kustomize-external-go-plugin:
	./hack/buildExternalGoPlugins.sh ./plugin

.PHONY: clean-kustomize-external-go-plugin
clean-kustomize-external-go-plugin:
	./hack/buildExternalGoPlugins.sh ./plugin clean
