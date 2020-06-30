package main

import (
	"testing"
)

func TestVersionFromAndToString(t *testing.T) {
	vs := "1.1.1"
	expect := "v1.1.1"
	v, err := newModuleVersionFromString(vs)
	if err != nil {
		t.Error(err)
	}
	if v.String() != expect {
		t.Errorf("%s doesn't match expected %s", v.String(), expect)
	}

	vs = "v0.0.0"
	expect = "v0.0.0"
	v, err = newModuleVersionFromString(vs)
	if err != nil {
		t.Error(err)
	}
	if v.String() != expect {
		t.Errorf("%s doesn't match expected %s", v.String(), expect)
	}

	vs = "v0.0"
	v, err = newModuleVersionFromString(vs)
	if err == nil {
		t.Errorf("%s should be invalid", vs)
	}

	vs = ""
	v, err = newModuleVersionFromString(vs)
	if err == nil {
		t.Errorf("%s should be invalid", vs)
	}
}

func TestVersionFromGitTags(t *testing.T) {
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
	expect := "v0.1.11"

	v, err := newModuleVersionFromGitTags(tags, "cmd/config")
	if err != nil {
		t.Error(err)
	}
	if v.String() != expect {
		t.Errorf("%s doesn't match expected %s", v.String(), expect)
	}
}

func TestVersionBumpPatch(t *testing.T) {
	v := moduleVersion{0, 1, 1}
	expect := "v0.1.2"
	err := v.Bump("patch")
	if err != nil {
		t.Error(err)
	}

	if v.String() != expect {
		t.Errorf("%s doesn't match expected %s", v.String(), expect)
	}
}

func TestVersionBumpMinor(t *testing.T) {
	v := moduleVersion{0, 1, 1}
	expect := "v0.2.0"
	err := v.Bump("minor")
	if err != nil {
		t.Error(err)
	}

	if v.String() != expect {
		t.Errorf("%s doesn't match expected %s", v.String(), expect)
	}
}

func TestVersionBumpMajor(t *testing.T) {
	v := moduleVersion{0, 1, 1}
	expect := "v1.0.0"
	err := v.Bump("major")
	if err != nil {
		t.Error(err)
	}

	if v.String() != expect {
		t.Errorf("%s doesn't match expected %s", v.String(), expect)
	}
}

func TestVersionBumpError(t *testing.T) {
	v := moduleVersion{}
	err := v.Bump("unknown")
	if err == nil {
		t.Errorf("Invalid bumping type should have error")
	}
}
