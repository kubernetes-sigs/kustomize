package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestModuleTags(t *testing.T) {
	tags := `api/v0.1.1
	api/v0.2.0
	api/v0.3.0
	api/v0.3.1
	api/v0.3.2
	api/v0.3.3
	cmd/config/v0.0.1
	cmd/config/v0.0.10
	cmd/config/v0.0.11
	cmd/config/v0.0.12
	cmd/config/v0.0.13
	cmd/config/v0.0.2
	cmd/config/v0.0.3
	cmd/config/v0.0.4
	cmd/config/v0.0.5
	cmd/config/v0.0.6
	cmd/config/v0.0.7
	cmd/config/v0.0.8
	cmd/config/v0.0.9
	cmd/config/v0.1.0
	cmd/config/v0.1.1
	cmd/config/v0.1.10
	cmd/config/v0.1.11
	cmd/config/v0.1.2
	cmd/config/v0.1.3
	cmd/config/v0.1.4
	cmd/config/v0.1.5
	cmd/config/v0.1.6
	cmd/config/v0.1.7
	cmd/config/v0.1.8
	cmd/kubectl/v0.0.1
	cmd/kubectl/v0.0.2
	cmd/kubectl/v0.0.3
	cmd/resource/v0.0.1
	cmd/resource/v0.0.2
	kustomize/v3.2.1
	kustomize/v3.2.2
	kustomize/v3.2.3
	kustomize/v3.3.0
	kustomize/v3.4.0
	kustomize/v3.5.1
	kustomize/v3.5.2
	kustomize/v3.5.3
	kustomize/v3.5.4
	kustomize/v3.5.5`
	expect := "cmd/config/v0.1.11"

	m := module{
		name: "cmd/config",
	}

	err := m.UpdateVersion(tags)
	if err != nil {
		t.Error(err)
	}

	if m.Tag() != expect {
		t.Errorf("Tag %s doesn't match expected %s", m.Tag(), expect)
	}
}

func TestCheckModReplace1(t *testing.T) {
	path := "testpath"
	dataString := `module sigs.k8s.io/kustomize/kustomize/v3

	go 1.13
	
	replace (
		sigs.k8s.io/kustomize/cmd/kubectl v0.0.3 => ../cmd/kubectl
	)`

	expect := `Found replace in testpath
 - Please update go.mod to pin a specific version of sigs.k8s.io/kustomize/cmd/kubectl
`

	err := checkModReplace(path, []byte(dataString))
	if err.Error() != expect {
		t.Errorf("Error %s doesn't match expected %s", err.Error(), expect)
	}
}

func TestCheckModReplace2(t *testing.T) {
	path := "testpath"
	dataString := `module sigs.k8s.io/kustomize/kustomize/v3

	go 1.13
	
	replace sigs.k8s.io/kustomize/cmd/kubectl v0.0.3 => ../cmd/kubectl`

	expect := `Found replace in testpath
 - Please update go.mod to pin a specific version of sigs.k8s.io/kustomize/cmd/kubectl
`

	err := checkModReplace(path, []byte(dataString))
	if err.Error() != expect {
		t.Errorf("Error %s doesn't match expected %s", err.Error(), expect)
	}
}

func TestCheckModReplace3(t *testing.T) {
	path := "testpath"
	dataString := `module sigs.k8s.io/kustomize/kustomize/v3

	go 1.13
	
	exclude (
		github.com/russross/blackfriday v2.0.0+incompatible
		sigs.k8s.io/kustomize/api v0.2.0
	)`

	err := checkModReplace(path, []byte(dataString))
	if err != nil {
		t.Errorf("Error %s is not expected", err.Error())
	}
}

func TestCheckModReplaceWithFile(t *testing.T) {
	dataString := `module sigs.k8s.io/kustomize/kustomize/v3

	go 1.13
	
	exclude (
		github.com/russross/blackfriday v2.0.0+incompatible
		sigs.k8s.io/kustomize/api v0.2.0
	)`

	dir, err := ioutil.TempDir("", "kustomize-releases-test")
	if err != nil {
		t.Error(err)
	}
	modName := "kustomize"
	defer os.RemoveAll(dir)

	err = os.MkdirAll(filepath.Join(dir, modName), os.FileMode(0700))
	if err != nil {
		t.Error(err)
	}

	ioutil.WriteFile(filepath.Join(dir, modName, "go.mod"), []byte(dataString), os.FileMode(0600))

	m := module{
		name: modName,
		path: dir,
	}

	err = m.CheckModReplace()
	if err != nil {
		t.Errorf("Error %s is not expected", err.Error())
	}
}
