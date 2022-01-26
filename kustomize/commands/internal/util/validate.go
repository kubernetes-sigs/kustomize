// Copyright 2022 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"errors"
	"fmt"
	"strings"

	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/internal/kustfile"
)

type BuildMetadataValidator struct{}

func (b *BuildMetadataValidator) Validate(args []string) ([]string, error) {
	if len(args) == 0 {
		return nil, errors.New("must specify a buildMetadata option")
	}
	if len(args) > 1 {
		return nil, fmt.Errorf("too many arguments: %s; to provide multiple buildMetadata options, please separate options by comma", args)
	}
	opts := strings.Split(args[0], ",")
	for _, opt := range opts {
		if !kustfile.StringInSlice(opt, types.BuildMetadataOptions) {
			return nil, fmt.Errorf("invalid buildMetadata option: %s", opt)
		}
	}
	return opts, nil
}
