package main

import (
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
	kstatus/v0.0.1
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
