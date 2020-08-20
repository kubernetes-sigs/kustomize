package main

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
)

func Test_executeKustomizeTestBuild(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	defer os.RemoveAll(tmpDir)

	kustomizationYamlFilePath := path.Join(tmpDir, "kustomization.yaml")
	kustomizationYaml := `
generatorOptions:
  disableNameSuffixHash: true
configMapGenerator:
- name: foo-config
  literals:    
  - foo=bar
`
	err = ioutil.WriteFile(kustomizationYamlFilePath, []byte(kustomizationYaml), os.ModePerm)
	if err != nil {
		t.Fatalf("error writing kustomization file to path: %v error: %v\n", kustomizationYamlFilePath, err)
	}

	result, err := executeKustomizeBuild(tmpDir)
	if err != nil {
		t.Fatalf("unexpected kustomize error: %v\n", err)
	}

	expectedK8sYaml := `apiVersion: v1
data:
  foo: bar
kind: ConfigMap
metadata:
  name: foo-config
`
	if string(result) != expectedK8sYaml {
		t.Fatalf("expected k8s yaml: [%v] but got: [%v]\n", expectedK8sYaml, string(result))
	}
}

//func Test_foo(t *testing.T) {
//	os.Setenv("qlik_HELM_REPO_USERNAME", "dlx")
//	os.Setenv("qlik_HELM_REPO_PASSWORD", "AKCp5btLBQMnoy8rjT2zTrcg6PXnFM6sAzXfTcV5hZDP9W9uFLHs374uidqDmrKkcy4waUnQd")
//	result, err := executeKustomizeBuild("/Users/dlx/node/elastic-charts/develop-QCS")
//	if err != nil {
//		t.Fatalf("unexpected kustomize error: %v\n", err)
//	}
//	fmt.Println(string(result))
//}

func executeKustomizeBuild(directory string) ([]byte, error) {
	kustomizer := krusty.MakeKustomizer(filesys.MakeFsOnDisk(), &krusty.Options{
		DoLegacyResourceSort: false,
		LoadRestrictions:     types.LoadRestrictionsNone,
		DoPrune:              false,
		PluginConfig:         konfig.DisabledPluginConfig(),
	})
	resMap, err := kustomizer.Run(directory)
	if err != nil {
		return nil, err
	}
	return resMap.AsYaml()
}
