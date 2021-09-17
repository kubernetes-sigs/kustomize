package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestVaultPlugin(t *testing.T) {
	RegisterFailHandler(Fail)         // registers the fail handler from ginkgo
	RunSpecs(t, "Testing Acceptance") // hands over control to the ginkgo testing framework
}

var _ = Describe("Testing acceptance", func() {

})
