// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package comments

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// deploymentsFixture returns n Deployments with labels, annotations, a containers list
// (matched by its merge key) with env lists, and a head or line comment on about every
// fifth field, so CopyComments walks a mix of commented and comment-free fields.
func deploymentsFixture(n int) string {
	var b strings.Builder
	field := 0
	// comment returns a line comment for about every fifth field.
	comment := func() string {
		field++
		if field%5 == 0 {
			return fmt.Sprintf(" # note %d", field)
		}
		return ""
	}
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteString("---\n")
		}
		fmt.Fprintf(&b, "# Deployment %d\n", i)
		b.WriteString("apiVersion: apps/v1\n")
		b.WriteString("kind: Deployment\n")
		b.WriteString("metadata:\n")
		fmt.Fprintf(&b, "  name: app-%d%s\n", i, comment())
		b.WriteString("  labels:\n")
		for l := 0; l < 4; l++ {
			fmt.Fprintf(&b, "    label-%d: value-%d%s\n", l, l, comment())
		}
		b.WriteString("  annotations:\n")
		for a := 0; a < 3; a++ {
			fmt.Fprintf(&b, "    example.com/annotation-%d: value-%d%s\n", a, a, comment())
		}
		b.WriteString("spec:\n")
		fmt.Fprintf(&b, "  replicas: %d%s\n", i%5+1, comment())
		b.WriteString("  template:\n")
		b.WriteString("    spec:\n")
		b.WriteString("      # containers are matched by name\n")
		b.WriteString("      containers:\n")
		for c := 0; c < 3; c++ {
			fmt.Fprintf(&b, "      - name: container-%d%s\n", c, comment())
			fmt.Fprintf(&b, "        image: registry.example.com/image-%d:v%d%s\n", c, i, comment())
			b.WriteString("        env:\n")
			for e := 0; e < 4; e++ {
				fmt.Fprintf(&b, "        - name: ENV_%d%s\n", e, comment())
				fmt.Fprintf(&b, "          value: \"%d\"\n", e)
			}
			b.WriteString("        ports:\n")
			fmt.Fprintf(&b, "        - containerPort: %d%s\n", 8080+c, comment())
		}
	}
	return b.String()
}

func readFixture(t testing.TB, n int) []*yaml.RNode {
	t.Helper()
	nodes, err := (&kio.ByteReader{
		Reader:                strings.NewReader(deploymentsFixture(n)),
		OmitReaderAnnotations: true,
	}).Read()
	require.NoError(t, err)
	return nodes
}

func clearComments(n *yaml.Node) {
	n.HeadComment, n.LineComment, n.FootComment = "", "", ""
	for _, c := range n.Content {
		clearComments(c)
	}
}

// strippedCopies returns copies of nodes with every comment removed, as after a
// function that drops comments.
func strippedCopies(nodes []*yaml.RNode) []*yaml.RNode {
	out := make([]*yaml.RNode, len(nodes))
	for i := range nodes {
		out[i] = nodes[i].Copy()
		clearComments(out[i].YNode())
	}
	return out
}

// Restoring comments onto comment-stripped copies must reproduce the original exactly.
func TestCopyCommentsRestoresStrippedFixture(t *testing.T) {
	orig := readFixture(t, 20)
	restored := strippedCopies(orig)
	for i := range orig {
		require.NoError(t, CopyComments(orig[i], restored[i]))
		want, err := orig[i].String()
		require.NoError(t, err)
		got, err := restored[i].String()
		require.NoError(t, err)
		assert.Equal(t, want, got, "resource %d", i)
	}
}

// BenchmarkCopyComments restores comments onto comment-stripped copies of 200
// Deployments, as after each function that drops comments.
func BenchmarkCopyComments(b *testing.B) {
	orig := readFixture(b, 200)
	b.ReportAllocs()
	for b.Loop() {
		b.StopTimer()
		restored := strippedCopies(orig)
		b.StartTimer()
		for i := range orig {
			if err := CopyComments(orig[i], restored[i]); err != nil {
				b.Fatal(err)
			}
		}
	}
}
