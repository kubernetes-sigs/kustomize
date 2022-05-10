package provenance_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/provenance"
)

func TestGetProvenance(t *testing.T) {
	p := provenance.GetProvenance()
	// These are not set during go test: https://github.com/golang/go/issues/33976
	assert.Equal(t, "unknown", p.Version)
	assert.Equal(t, "unknown", p.BuildDate)
	assert.Equal(t, "unknown", p.GitCommit)
	assert.Equal(t, "unknown", p.GitTreeState)

	// These are set properly during go test
	assert.NotEmpty(t, p.GoArch)
	assert.NotEmpty(t, p.GoOs)
	assert.Contains(t, p.GoVersion, "go1.")
}

func TestProvenance_Short(t *testing.T) {
	p := provenance.GetProvenance()
	// The version not set during go test, so this comes from an ldflag: https://github.com/golang/go/issues/33976
	assert.Equal(t, "unknown", p.Short())

	p.Version = "kustomize/v4.11.12"
	assert.Equal(t, "v4.11.12", p.Short())
}

func TestProvenance_Full(t *testing.T) {
	p := provenance.GetProvenance()
	// Most values are not set during go test: https://github.com/golang/go/issues/33976
	assert.Contains(t, p.Full(), "{Version:unknown GitCommit:unknown GitTreeState:unknown BuildDate:unknown")
	assert.Regexp(t, "GoOs:\\w+ GoArch:\\w+ GoVersion:go1", p.Full())
}

func TestProvenance_Semver(t *testing.T) {
	p := provenance.GetProvenance()
	// The version not set during go test
	assert.Equal(t, "unknown", p.Semver())

	p.Version = "kustomize/v4.11.12"
	assert.Equal(t, "v4.11.12", p.Semver())
}
