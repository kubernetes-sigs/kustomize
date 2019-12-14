// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Code generated by "stringer -type=pluginType"; DO NOT EDIT.

package main

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[unknown-0]
	_ = x[Transformer-1]
	_ = x[Generator-2]
}

const _pluginType_name = "unknownTransformerGenerator"

var _pluginType_index = [...]uint8{0, 7, 18, 27}

func (i pluginType) String() string {
	if i < 0 || i >= pluginType(len(_pluginType_index)-1) {
		return "pluginType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _pluginType_name[_pluginType_index[i]:_pluginType_index[i+1]]
}
