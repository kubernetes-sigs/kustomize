// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package set

import (
	"fmt"
	"strings"
	"testing"

	testutils_test "sigs.k8s.io/kustomize/kustomize/v4/commands/internal/testutils"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func TestSetImage(t *testing.T) {
	type given struct {
		args         []string
		infileImages []string
	}
	type expected struct {
		fileOutput []string
		err        error
	}
	testCases := []struct {
		description string
		given       given
		expected    expected
	}{
		{
			given: given{
				args: []string{"image1=my-image1:my-tag"},
			},
			expected: expected{
				fileOutput: []string{
					"images:",
					"- name: image1",
					"  newName: my-image1",
					"  newTag: my-tag",
				}},
		},
		{
			given: given{
				args: []string{"image1=my-image1:1234"},
			},
			expected: expected{
				fileOutput: []string{
					"images:",
					"- name: image1",
					"  newName: my-image1",
					"  newTag: \"1234\"",
				}},
		},
		{
			given: given{
				args: []string{"image1=my-image1:3.2e2"},
			},
			expected: expected{
				fileOutput: []string{
					"images:",
					"- name: image1",
					"  newName: my-image1",
					"  newTag: \"3.2e2\"",
				}},
		},
		{
			given: given{
				args: []string{"image1=my-image1@sha256:24a0c4b4a4c0eb97a1aabb8e29f18e917d05abfe1b7a7c07857230879ce7d3d3"},
			},
			expected: expected{
				fileOutput: []string{
					"images:",
					"- digest: sha256:24a0c4b4a4c0eb97a1aabb8e29f18e917d05abfe1b7a7c07857230879ce7d3d3",
					"  name: image1",
					"  newName: my-image1",
				}},
		},
		{
			given: given{
				args: []string{"image1:my-tag"},
			},
			expected: expected{
				fileOutput: []string{
					"images:",
					"- name: image1",
					"  newTag: my-tag",
				}},
		},
		{
			given: given{
				args: []string{"image1@sha256:24a0c4b4a4c0eb97a1aabb8e29f18e917d05abfe1b7a7c07857230879ce7d3d3"},
			},
			expected: expected{
				fileOutput: []string{
					"images:",
					"- digest: sha256:24a0c4b4a4c0eb97a1aabb8e29f18e917d05abfe1b7a7c07857230879ce7d3d3",
					"  name: image1",
				}},
		},
		{
			description: "image with tag and digest",
			given: given{
				args: []string{"org/image1:tag@sha256:24a0c4b4a4c0eb97a1aabb8e29f18e917d05abfe1b7a7c07857230879ce7d3d3"},
			},
			expected: expected{
				fileOutput: []string{
					"images:",
					"- digest: sha256:24a0c4b4a4c0eb97a1aabb8e29f18e917d05abfe1b7a7c07857230879ce7d3d3",
					"  name: org/image1",
					"  newTag: tag",
				}},
		},
		{
			description: "<image>=<image>",
			given: given{
				args: []string{"ngnix=localhost:5000/my-project/ngnix"},
			},
			expected: expected{
				fileOutput: []string{
					"images:",
					"- name: ngnix",
					"  newName: localhost:5000/my-project/ngnix",
				}},
		},
		{
			given: given{
				args: []string{"ngnix=localhost:5000/my-project/ngnix:dev-01"},
			},
			expected: expected{
				fileOutput: []string{
					"images:",
					"- name: ngnix",
					"  newName: localhost:5000/my-project/ngnix",
					"  newTag: dev-01",
				}},
		},
		{
			description: "override file",
			given: given{
				args: []string{"image1=foo.bar.foo:8800/foo/image1:foo-bar"},
				infileImages: []string{
					"images:",
					"- name: image1",
					"  newName: my-image1",
					"  newTag: my-tag",
					"- name: image2",
					"  newName: my-image2",
					"  newTag: my-tag2",
				},
			},
			expected: expected{
				fileOutput: []string{
					"images:",
					"- name: image1",
					"  newName: foo.bar.foo:8800/foo/image1",
					"  newTag: foo-bar",
					"- name: image2",
					"  newName: my-image2",
					"  newTag: my-tag2",
				}},
		},
		{
			description: "override file with patch",
			given: given{
				args: []string{"image1=foo.bar.foo:8800/foo/image1:foo-bar"},
				infileImages: []string{
					"images:",
					"- name: image1",
					"  newName: my-image1",
					"  newTag: my-tag",
					"- name: image2",
					"  newName: my-image2",
					"  newTag: my-tag2",
					"patchesJson6902:",
					"- patch: |-",
					"    - op: remove",
					"      path: /spec/selector",
					"  target:",
					"    kind: Service",
					"    name: foo",
					"    version: v1",
				},
			},
			expected: expected{
				fileOutput: []string{
					"images:",
					"- name: image1",
					"  newName: foo.bar.foo:8800/foo/image1",
					"  newTag: foo-bar",
					"- name: image2",
					"  newName: my-image2",
					"  newTag: my-tag2",
					"patchesJson6902:",
					"- patch: |-",
					"    - op: remove",
					"      path: /spec/selector",
					"  target:",
					"    kind: Service",
					"    name: foo",
					"    version: v1",
				}},
		},
		{
			description: "override new tag and new name with just a new tag",
			given: given{
				args: []string{"image1:v1"},
				infileImages: []string{
					"images:",
					"- name: image1",
					"  newTag: my-tag",
				},
			},
			expected: expected{
				fileOutput: []string{
					"images:",
					"- name: image1",
					"  newTag: v1",
				}},
		},
		{
			description: "multiple args with multiple overrides",
			given: given{
				args: []string{
					"image1=foo.bar.foo:8800/foo/image1:foo-bar",
					"image2=my-image2@sha256:24a0c4b4a4c0eb97a1aabb8e29f18e917d05abfe1b7a7c07857230879ce7d3d3",
					"image3:my-tag",
				},
				infileImages: []string{
					"images:",
					"- name: image1",
					"  newName: my-image1",
					"  newTag: my-tag1",
					"- name: image2",
					"  newName: my-image2",
					"  newTag: my-tag2",
					"- name: image3",
					"  newTag: my-tag",
				},
			},
			expected: expected{
				fileOutput: []string{
					"images:",
					"- name: image1",
					"  newName: foo.bar.foo:8800/foo/image1",
					"  newTag: foo-bar",
					"- digest: sha256:24a0c4b4a4c0eb97a1aabb8e29f18e917d05abfe1b7a7c07857230879ce7d3d3",
					"  name: image2",
					"  newName: my-image2",
					"- name: image3",
					"  newTag: my-tag",
				}},
		},
		{
			description: "error: no args",
			expected: expected{
				err: errImageNoArgs,
			},
		},
		{
			description: "error: invalid args",
			given: given{
				args: []string{"bad", "args"},
			},
			expected: expected{
				err: errImageInvalidArgs,
			},
		},
		{
			description: "override new tag but keep new name",
			given: given{
				args: []string{"image1=*:v1"},
				infileImages: []string{
					"images:",
					"- name: image1",
					"  newName: foo.bar.foo:8800/foo/image1",
					"  newTag: my-tag",
				},
			},
			expected: expected{
				fileOutput: []string{
					"images:",
					"- name: image1",
					"  newName: foo.bar.foo:8800/foo/image1",
					"  newTag: v1",
				}},
		},
		{
			description: "override new name but keep new tag",
			given: given{
				args: []string{"image1=my-image1:*"},
				infileImages: []string{
					"images:",
					"- name: image1",
					"  newName: foo.bar.foo:8800/foo/image1",
					"  newTag: my-tag",
				},
			},
			expected: expected{
				fileOutput: []string{
					"images:",
					"- name: image1",
					"  newName: my-image1",
					"  newTag: my-tag",
				}},
		},
		{
			description: "keep new name and new tag (rare case)",
			given: given{
				args: []string{"image1=*:*"},
				infileImages: []string{
					"images:",
					"- name: image1",
					"  newName: my-image1",
					"  newTag: my-tag",
				},
			},
			expected: expected{
				fileOutput: []string{
					"images:",
					"- name: image1",
					"  newName: my-image1",
					"  newTag: my-tag",
				}},
		},
		{
			description: "do not set asterisk as new name for existing image",
			given: given{
				args: []string{"image1=*:v1"},
				infileImages: []string{
					"images:",
					"- name: image1",
					"  newTag: my-tag",
				},
			},
			expected: expected{
				fileOutput: []string{
					"images:",
					"- name: image1",
					"  newTag: v1",
				}},
		},
		{
			description: "do not set asterisk as new name",
			given: given{
				args: []string{"image1=*:v1"},
			},
			expected: expected{
				fileOutput: []string{
					"images:",
					"- name: image1",
					"  newTag: v1",
				}},
		},
		{
			description: "do not set asterisk as new tag",
			given: given{
				args: []string{"image1=my-image1:*"},
			},
			expected: expected{
				fileOutput: []string{
					"images:",
					"- name: image1",
					"  newName: my-image1",
				}},
		},
		{
			description: "keep new name and update digest",
			given: given{
				args: []string{"image2=*@sha256:24a0c4b4a4c0eb97a1aabb8e29f18e917d05abfe1b7a7c07857230879ce7d3d3"},
				infileImages: []string{
					"images:",
					"- name: image2",
					"  newName: my-image2",
					"  digest: sha256:abcdef12345",
				},
			},
			expected: expected{
				fileOutput: []string{
					"images:",
					"- digest: sha256:24a0c4b4a4c0eb97a1aabb8e29f18e917d05abfe1b7a7c07857230879ce7d3d3",
					"  name: image2",
					"  newName: my-image2",
				}},
		},
		{
			description: "keep new name, remove tag, and set digest",
			given: given{
				args: []string{"image2=*@sha256:24a0c4b4a4c0eb97a1aabb8e29f18e917d05abfe1b7a7c07857230879ce7d3d3"},
				infileImages: []string{
					"images:",
					"- name: image2",
					"  newName: my-image2",
					"  newTag: my-tag",
				},
			},
			expected: expected{
				fileOutput: []string{
					"images:",
					"- digest: sha256:24a0c4b4a4c0eb97a1aabb8e29f18e917d05abfe1b7a7c07857230879ce7d3d3",
					"  name: image2",
					"  newName: my-image2",
				}},
		},
		{
			description: "update new name and keep the digest",
			given: given{
				args: []string{"image2=my-image2@*"},
				infileImages: []string{
					"images:",
					"- digest: sha256:24a0c4b4a4c0eb97a1aabb8e29f18e917d05abfe1b7a7c07857230879ce7d3d3",
					"  name: image2",
					"  newName: foo.bar.foo:8800/foo/image1",
				},
			},
			expected: expected{
				fileOutput: []string{
					"images:",
					"- digest: sha256:24a0c4b4a4c0eb97a1aabb8e29f18e917d05abfe1b7a7c07857230879ce7d3d3",
					"  name: image2",
					"  newName: my-image2",
				}},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s%v", tc.description, tc.given.args), func(t *testing.T) {
			fSys := filesys.MakeFsInMemory()
			cmd := newCmdSetImage(fSys)

			if len(tc.given.infileImages) > 0 {
				// write file with infileImages
				testutils_test.WriteTestKustomizationWith(
					fSys,
					[]byte(strings.Join(tc.given.infileImages, "\n")))
			} else {
				testutils_test.WriteTestKustomization(fSys)
			}

			// act
			err := cmd.RunE(cmd, tc.given.args)

			// assert
			if err != tc.expected.err {
				t.Errorf("Unexpected error from set image command. Actual: %v\nExpected: %v", err, tc.expected.err)
				t.FailNow()
			}

			content, err := testutils_test.ReadTestKustomization(fSys)
			if err != nil {
				t.Errorf("unexpected read error: %v", err)
				t.FailNow()
			}
			expectedStr := strings.Join(tc.expected.fileOutput, "\n")
			if !strings.Contains(string(content), expectedStr) {
				t.Errorf("unexpected image in kustomization file. \nActual:\n%s\nExpected:\n%s", content, expectedStr)
			}
		})
	}
}
