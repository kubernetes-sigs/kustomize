// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package resource_test

import (
	"flag"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Register custom flags before parsing
	flag.Bool("update-golden", false, "update golden files for resource tests")
	// Parse flags to register custom flags like -update-golden
	flag.Parse()
	os.Exit(m.Run())
}
