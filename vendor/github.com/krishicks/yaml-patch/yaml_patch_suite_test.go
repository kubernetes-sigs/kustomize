package yamlpatch_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestYamlPatch(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "YamlPatch Suite")
}
