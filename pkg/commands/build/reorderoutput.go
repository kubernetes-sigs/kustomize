// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"fmt"

	"github.com/spf13/pflag"
)

//go:generate stringer -type=reorderOutput
type reorderOutput int

const (
	unspecified reorderOutput = iota
	none
	legacy
	kubectlapply
	kubectldelete
)

const (
	flagReorderOutputName = "reorder"
)

var (
	flagReorderOutputValue = legacy.String()
	flagReorderOutputHelp  = "Reorder the resources just before output. " +
		"Use '" + legacy.String() + "' to apply a legacy reordering (Namespaces first, Webhooks last, etc). " +
		"Use '" + kubectlapply.String() + "' to apply a kubectl apply friendly reordering (Namespaces first, etc). " +
		"Use '" + kubectldelete.String() + "' to apply a kubetl delete friendy reordering (Namespaces last, etc). " +
		"Use '" + none.String() + "' to suppress a final reordering."
)

func addFlagReorderOutput(set *pflag.FlagSet) {
	set.StringVar(
		&flagReorderOutputValue, flagReorderOutputName,
		legacy.String(), flagReorderOutputHelp)
}

func validateFlagReorderOutput() (reorderOutput, error) {
	switch flagReorderOutputValue {
	case none.String():
		return none, nil
	case legacy.String():
		return legacy, nil
	case kubectlapply.String():
		return kubectlapply, nil
	case kubectldelete.String():
		return kubectldelete, nil
	default:
		return unspecified, fmt.Errorf(
			"illegal flag value --%s %s; legal values: %v",
			flagReorderOutputName, flagReorderOutputValue,
			[]string{legacy.String(), kubectlapply.String(), kubectldelete.String(), none.String()})
	}
}
