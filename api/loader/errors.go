// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package loader

import "fmt"

var (
	ErrHTTP   = fmt.Errorf("HTTP Error")
	ErrLdrDir = fmt.Errorf("can only create loader at directory")
)
