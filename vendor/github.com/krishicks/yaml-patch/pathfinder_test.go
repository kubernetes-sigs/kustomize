package yamlpatch_test

import (
	yamlpatch "github.com/krishicks/yaml-patch"
	yaml "gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Pathfinder", func() {
	var pathfinder *yamlpatch.PathFinder

	BeforeEach(func() {
		var iface interface{}

		bs := []byte(`
jobs:
- name: job1
  plan:
  - get: A
    args:
    - arg: arg1
    - arg: arg2
    bool: true
  - get: B
  - get: C/D

- name: job2
  plan:
  - aggregate:
    - get: C
    - get: A
`)

		err := yaml.Unmarshal(bs, &iface)
		Expect(err).NotTo(HaveOccurred())
		container := yamlpatch.NewNode(&iface).Container()
		pathfinder = yamlpatch.NewPathFinder(container)
	})

	Describe("Find", func() {
		DescribeTable(
			"should",
			func(path string, expected []string) {
				actual := pathfinder.Find(path)
				Expect(actual).To(HaveLen(len(expected)))
				for _, el := range expected {
					Expect(actual).To(ContainElement(el))
				}
			},
			Entry("return a route for the root object", "/", []string{"/"}),
			Entry("return a route for an object under the root", "/jobs", []string{"/jobs"}),
			Entry("return a route for an element within an object under the root", "/jobs/0", []string{"/jobs/0"}),
			Entry("return a route for an object within an element within an object under the root", "/jobs/0/plan", []string{"/jobs/0/plan"}),
			Entry("return a route for an object within an element within an object under the root", "/jobs/0/plan/1", []string{"/jobs/0/plan/1"}),
			Entry("return routes for multiple matches", "/jobs/get=A", []string{"/jobs/0/plan/0", "/jobs/1/plan/0/aggregate/1"}),
			Entry("return a route for a single submatch with help", "/jobs/get=A/args/arg=arg2", []string{"/jobs/0/plan/0/args/1"}),
			Entry("return a route for a single submatch with no help", "/jobs/get=A/arg=arg2", []string{"/jobs/0/plan/0/args/1"}),
			Entry("return a route for a single submatch with help using escape ordering", "/jobs/get=C~1D", []string{"/jobs/0/plan/2"}),
			Entry("return a route when given a pointer with a leaf that does not exist", "/jobs/name=job1/nonexistent", []string{"/jobs/0/nonexistent"}),
			Entry("return a route when given a pointer with an array thingy", "/jobs/name=job1/plan/-", []string{"/jobs/0/plan/-"}),
		)
		DescribeTable(
			"should not",
			func(path string) {
				Expect(pathfinder.Find(path)).To(BeNil())
			},
			Entry("return any routes when given a bad index", "/jobs/2"),
			Entry("return any routes when given a bad index", "/jobs/-1"),
		)
	})
})
