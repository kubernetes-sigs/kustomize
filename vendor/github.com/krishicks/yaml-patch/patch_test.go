package yamlpatch_test

import (
	yamlpatch "github.com/krishicks/yaml-patch"
	yaml "gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Patch", func() {
	Describe("Apply", func() {
		DescribeTable(
			"positive cases",
			func(doc, ops, expectedYAML string) {
				patch, err := yamlpatch.DecodePatch([]byte(ops))
				Expect(err).NotTo(HaveOccurred())

				actualBytes, err := patch.Apply([]byte(doc))
				Expect(err).NotTo(HaveOccurred())

				var actualIface interface{}
				err = yaml.Unmarshal(actualBytes, &actualIface)
				Expect(err).NotTo(HaveOccurred())

				var expectedIface interface{}
				err = yaml.Unmarshal([]byte(expectedYAML), &expectedIface)
				Expect(err).NotTo(HaveOccurred())

				Expect(actualIface).To(Equal(expectedIface))
			},
			Entry("adding an element to an object",
				`---
foo: bar
`,
				`---
- op: add
  path: /baz
  value: qux
`,
				`---
foo: bar
baz: qux
`,
			),
			Entry("adding an element to an array",
				`---
foo: [bar,baz]
`,
				`---
- op: add
  path: /foo/1
  value: qux
`,
				`---
foo: [bar,qux,baz]
`,
			),
			Entry("removing an element from an object",
				`---
foo: bar
baz: qux
`,
				`---
- op: remove
  path: /baz
`,
				`---
foo: bar
`,
			),
			Entry("removing an element from an array",
				`---
foo: [bar,qux,baz]
`,
				`---
- op: remove
  path: /foo/1
`,
				`---
foo: [bar,baz]
`,
			),
			Entry("replacing an element in an object",
				`---
foo: bar
baz: qux
`,
				`---
- op: replace
  path: /baz
  value: boo
`,
				`---
foo: bar
baz: boo
`,
			),
			Entry("moving an element in an object",
				`---
foo:
  bar: baz
  waldo: fred
qux:
  corge: grault
`,
				`---
- op: move
  from: /foo/waldo
  path: /qux/thud
`,
				`---
foo:
  bar: baz
qux:
  corge: grault
  thud: fred
`,
			),
			Entry("moving an element in an array",
				`---
foo: [all, grass, cows, eat]
`,
				`---
- op: move
  from: /foo/1
  path: /foo/3
`,
				`---
foo: [all, cows, eat, grass]
`,
			),
			Entry("adding an object to an object",
				`---
foo: bar
`,
				`---
- op: add
  path: /child
  value:
    grandchild: {}
`,
				`---
foo: bar
child:
  grandchild: {}
`,
			),
			Entry("appending an element to an array",
				`---
foo: [bar]
`,
				`---
- op: add
  path: /foo/-
  value: [abc,def]
`,
				`---
foo: [bar, [abc, def]]
`,
			),
			Entry("removing a nil element from an object",
				`---
foo: bar
qux:
  baz: 1
  bar: ~
`,
				`---
- op: remove
  path: /qux/bar
`,
				`---
foo: bar
qux:
  baz: 1
`,
			),
			Entry("adding a nil element to an object",
				`---
foo: bar
`,
				`---
- op: add
  path: /baz
  value: ~
`,
				`---
foo: bar
baz: ~
`,
			),
			Entry("replacing the sole element in an array",
				`---
foo: [bar]
`,
				`---
- op: replace
  path: /foo/0
  value: baz
`,
				`---
foo: [baz]
`,
			),
			Entry("replacing an element in an array within a root array",
				`---
- foo: [bar, qux, baz]
`,
				`---
- op: replace
  path: /0/foo/0
  value: bum
`,
				`---
- foo: [bum, qux, baz]
`,
			),
			Entry("copying an element in an array within a root array with an index",
				`---
- foo: [bar, qux, baz]
  bar: [qux, baz]
`,
				`---
- op: copy
  from: /0/foo/0
  path: /0/bar/0
`,
				`---
- foo: [bar, qux, baz]
  bar: [bar, baz]
`,
			),
			Entry("testing for the existence of a nil value in an object",
				`---
baz: ~
`,
				`---
- op: test
  path: /baz
  value: ~
`,
				`---
baz: ~
`,
			),
			Entry("testing for the existence of a nil key in an object",
				`---
baz: ~
`,
				`---
- op: test
  path: /foo
  value: ~
`,
				`---
baz: ~
`,
			),
			Entry("testing for the existence of a nil value in an object",
				`---
baz: qux
`,
				`---
- op: test
  path: /baz
  value: qux
`,
				`---
baz: qux
`,
			),
			Entry("testing for the existence of an element in an array",
				`---
foo: [a, 2, c]
`,
				`---
- op: test
  path: /foo/1
  value: 2
`,
				`---
foo: [a, 2, c]
`,
			),
			Entry("testing for the existence of an element in an object using escape ordering",
				`---
baz/foo: qux
`,
				`---
- op: test
  path: /baz~1foo
  value: qux
`,
				`---
baz/foo: qux
`,
			),
			Entry("testing for the existence of an object",
				`---
foo:
  - bar: baz
    qux: corge
`,
				`---
- op: test
  path: /foo/0
  value:
    bar: baz
    qux: corge
`,
				`---
foo:
  - bar: baz
    qux: corge
`,
			),
			XEntry("copying an element in an array within a root array to a destination without an index",
				// this is in jsonpatch, but I'd like confirmation from the spec that this is intended
				`---
- foo: [bar, qux, baz]
  bar: [qux, baz]
`,
				`---
- op: copy
  from: /0/foo/0
  path: /0/bar
`,
				`---
- foo: [bar, qux, baz]
  bar: [bar, qux, baz]
`,
			),
			XEntry("copying an element in an array within a root array to a destination without an index",
				`---
- foo:
  bar: [qux, baz]
  baz:
    qux: bum
`,
				`---
- op: copy
  from: /0/foo/bar
  path: /0/baz/bar
`,
				`---
- foo: [bar, qux, baz]
  bar: [bar, qux, baz]
`,
			),
		)

		DescribeTable(
			"with extended syntax",
			func(doc, ops, expectedYAML string) {
				patch, err := yamlpatch.DecodePatch([]byte(ops))
				Expect(err).NotTo(HaveOccurred())

				actualBytes, err := patch.Apply([]byte(doc))
				Expect(err).NotTo(HaveOccurred())

				var actualIface interface{}
				err = yaml.Unmarshal(actualBytes, &actualIface)
				Expect(err).NotTo(HaveOccurred())

				var expectedIface interface{}
				err = yaml.Unmarshal([]byte(expectedYAML), &expectedIface)
				Expect(err).NotTo(HaveOccurred())

				Expect(actualIface).To(Equal(expectedIface))
			},
			Entry("a path that begins with a composite key",
				`---
- foo: bar
`,
				`---
- op: replace
  path: /foo=bar
  value:
    baz: quux
`,
				`---
- baz: quux
`,
			),
			Entry("a path that begins with an array index and ends with a composite key",
				`---
- waldo:
    - thud: boo
    - foo: bar
- corge: grault
`,
				`---
- op: replace
  path: /0/foo=bar
  value:
    baz: quux
`,
				`---
- waldo:
    - thud: boo
    - baz: quux
- corge: grault
`,
			),
			Entry("a path that begins with an object key and ends with a composite key",
				`---
waldo:
  - thud: boo
  - foo: bar
corge: grault
`,
				`---
- op: replace
  path: /waldo/foo=bar
  value:
    baz: quux
`,
				`---
waldo:
  - thud: boo
  - baz: quux
corge: grault
`,
			),
			Entry("a path that doesn't end with a composite key",
				`---
jobs:
- name: upgrade-opsmgr
  serial: true
  plan:
  - get: pivnet-opsmgr
  - put: something-else
`,
				`---
- op: replace
  path: /jobs/name=upgrade-opsmgr/plan/1
  value:
    get: something-else
`,
				`---
jobs:
- name: upgrade-opsmgr
  serial: true
  plan:
  - get: pivnet-opsmgr
  - get: something-else
`,
			),
			Entry("removes multiple entries in a single op",
				`---
foo:
  - bar: baz
  - waldo: fred
qux:
  corge: grault
  thud:
    - waldo: fred
    - bar: baz
`,
				`---
- op: remove
  path: /waldo=fred
`,
				`---
foo:
  - bar: baz
qux:
  corge: grault
  thud:
    - bar: baz
`,
			),
		)

		DescribeTable(
			"failure cases",
			func(doc, ops string) {
				patch, err := yamlpatch.DecodePatch([]byte(ops))
				Expect(err).NotTo(HaveOccurred())

				_, err = patch.Apply([]byte(doc))
				Expect(err).To(HaveOccurred())
			},
			Entry("adding an element to an object with a bad pointer",
				`---
foo: bar
`,
				`---
- op: add
  path: /baz/bat
  value: qux
`,
			),
			Entry("removing an element from an object with a bad pointer",
				`---
a:
  b:
    d: 1
`,
				`---
- op: remove
  path: /a/b/c
`,
			),
			Entry("moving an element in an object with a bad pointer",
				`---
a:
  b:
    d: 1
`,
				`---
- op: move
  from: /a/b/c
  path: /a/b/e
`,
			),
			Entry("removing an element from an array with a bad pointer",
				`---
a:
  b: [1]
`,
				`---
- op: remove
  path: /a/b/1
`,
			),
			Entry("moving an element from an array with a bad pointer",
				`---
a:
  b: [1]
`,
				`---
- op: move
  from: /a/b/1
  path: /a/b/2
`,
			),
			Entry("an operation with an invalid pathz field",
				`---
foo: bar
`,
				`---
- op: add
  pathz: /baz
  value: qux
`,
			),
			Entry("an add operation with an empty path",
				`---
foo: bar
`,
				`---
- op: add
  path: ''
  value: qux
`,
			),
			Entry("a replace operation on an array with an invalid path",
				`---
name:
  foo:
    bat
  qux:
    bum
`,
				`---
- op: replace
  path: /foo/2
  value: bum
`,
			),
		)
	})

	Describe("DecodePatch", func() {
		It("returns an empty patch when given nil", func() {
			patch, err := yamlpatch.DecodePatch(nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(patch).To(HaveLen(0))
		})

		It("returns a patch with a single op when given a single op", func() {
			ops := []byte(
				`---
- op: add
  path: /baz
  value: qux`)

			patch, err := yamlpatch.DecodePatch(ops)
			Expect(err).NotTo(HaveOccurred())

			var v interface{} = "qux"
			value := yamlpatch.NewNode(&v)
			Expect(patch).To(Equal(yamlpatch.Patch{
				{
					Op:    "add",
					Path:  "/baz",
					Value: value,
				},
			}))
		})
	})
})
