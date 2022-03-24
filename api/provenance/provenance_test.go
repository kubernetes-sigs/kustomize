package provenance_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/provenance"
)

func TestGetProvenance(t *testing.T) {
	p := provenance.GetProvenance()
	// These are not set during go test: https://github.com/golang/go/issues/33976
	// The values are therefore coming from the constants, which can be populated from ldflags during real builds
	assert.Equal(t, "v444.333.222", p.Version)
	assert.Equal(t, "1970-01-01T00:00:00Z", p.BuildDate)
	assert.Equal(t, "$Format:%H$", p.GitCommit)
	assert.Equal(t, "unknown", p.GitTreeState)

	// These are set properly during go test
	assert.NotEmpty(t, p.GoArch)
	assert.NotEmpty(t, p.GoOs)
	assert.Contains(t, p.GoVersion, "go1.")
}

func TestProvenance_Short(t *testing.T) {
	p := provenance.GetProvenance()
	// The version not set during go test, so this comes from an ldflag: https://github.com/golang/go/issues/33976
	assert.Equal(t, "v444.333.222", p.Short())

	p.Version = "kustomize/v4.11.12"
	assert.Equal(t, "v4.11.12", p.Semver())
}

func TestProvenance_Full(t *testing.T) {
	p := provenance.GetProvenance()
	// Most values are not set during go test: https://github.com/golang/go/issues/33976
	// version is populated from an ldflag
	assert.Contains(t, p.Full(), "{Version:v444.333.222 GitCommit:$Format:%H$ GitTreeState:unknown BuildDate:1970-01-01T00:00:00Z")
	assert.Regexp(t, "GoOs:\\w+ GoArch:\\w+ GoVersion:go1", p.Full())
}

func TestProvenance_Semver(t *testing.T) {
	p := provenance.GetProvenance()
	// The version not set during go test, so this comes from an ldflag: https://github.com/golang/go/issues/33976
	assert.Equal(t, "v444.333.222", p.Semver())
}
