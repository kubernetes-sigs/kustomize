// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kioutil_test

import (
	"bytes"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
)

func TestSortNodes_moreThan10(t *testing.T) {
	input := `
a: b
---
c: d
---
e: f
---
g: h
---
i: j
---
k: l
---
m: n
---
o: p
---
q: r
---
s: t
---
u: v
---
w: x
---
y: z
`
	actual := &bytes.Buffer{}
	rw := kio.ByteReadWriter{Reader: bytes.NewBufferString(input), Writer: actual}
	nodes, err := rw.Read()
	if !assert.NoError(t, err) {
		t.Fail()
	}

	// randomize the list
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(nodes), func(i, j int) { nodes[i], nodes[j] = nodes[j], nodes[i] })

	// sort them back into their original order
	if !assert.NoError(t, kioutil.SortNodes(nodes)) {
		t.Fail()
	}

	// check the sorted values
	expected := strings.Split(input, "---")
	for i := range nodes {
		a := strings.TrimSpace(nodes[i].MustString())
		b := strings.TrimSpace(expected[i])
		if !assert.Contains(t, a, b) {
			t.Fail()
		}
	}

	if !assert.NoError(t, rw.Write(nodes)) {
		t.Fail()
	}

	assert.Equal(t, strings.TrimSpace(input), strings.TrimSpace(actual.String()))
}
