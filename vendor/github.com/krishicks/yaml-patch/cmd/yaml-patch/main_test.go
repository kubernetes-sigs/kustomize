package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("yaml-patch", func() {
	It("builds", func() {
		_, err := gexec.Build("github.com/krishicks/yaml-patch/cmd/yaml-patch")
		Expect(err).NotTo(HaveOccurred())
	})
})
