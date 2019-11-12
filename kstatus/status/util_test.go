// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package status

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var testObj = map[string]interface{}{
	"f1": map[string]interface{}{
		"f2": map[string]interface{}{
			"i32":   int32(32),
			"i64":   int64(64),
			"float": 64.02,
			"ms": []interface{}{
				map[string]interface{}{"f1f2ms0f1": 22},
				map[string]interface{}{"f1f2ms1f1": "index1"},
			},
			"msbad": []interface{}{
				map[string]interface{}{"f1f2ms0f1": 22},
				32,
			},
		},
	},

	"ride": "dragon",

	"status": map[string]interface{}{
		"conditions": []interface{}{
			map[string]interface{}{"f1f2ms0f1": 22},
			map[string]interface{}{"f1f2ms1f1": "index1"},
		},
	},
}

func TestGetIntField(t *testing.T) {
	v := GetIntField(testObj, ".f1.f2.i32", -1)
	assert.Equal(t, int(32), v)

	v = GetIntField(testObj, ".f1.f2.wrongname", -1)
	assert.Equal(t, int(-1), v)

	v = GetIntField(testObj, ".f1.f2.i64", -1)
	assert.Equal(t, int(64), v)

	v = GetIntField(testObj, ".f1.f2.float", -1)
	assert.Equal(t, int(-1), v)
}

func TestGetStringField(t *testing.T) {
	v := GetStringField(testObj, ".ride", "horse")
	assert.Equal(t, v, "dragon")

	v = GetStringField(testObj, ".destination", "north")
	assert.Equal(t, v, "north")
}
