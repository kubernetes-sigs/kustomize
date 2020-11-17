#! /usr/bin/env bash

builtinPlugins=(AnnotationsTransformer \
	ConfigMapGenerator \
	HashTransformer \
	ImageTagTransformer \
	LabelTransformer \
	LegacyOrderTransformer \
	NamespaceTransformer \
	PatchJson6902Transformer \
	PatchStrategicMergeTransformer \
	PatchTransformer \
	PrefixSuffixTransformer \
	ReplicaCountTransformer \
	SecretGenerator \
	ValueAddTransformer \
	HelmChartInflationGenerator)

builtinPluginDir=../plugin/builtin

if [[ -z $KRM_FUNCTION_DIR ]]; then
    echo "Must specify output directory by \$KRM_FUNCTION_DIR"
    exit 1
fi


# Install pluginator
pushd ../cmd/pluginator
make install
popd


for pluginName in ${builtinPlugins[@]}; do
    dirName=$(echo $pluginName | tr '[:upper:]' '[:lower:]')
    srcPath="$builtinPluginDir/$dirName/$pluginName.go"
    dstPath="$KRM_FUNCTION_DIR/$dirName"
    pluginator krm -i $srcPath -o $dstPath
done
