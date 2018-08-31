package yamlpatch_test

import (
	yamlpatch "github.com/krishicks/yaml-patch"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PlaceholderWrapper", func() {
	var placeholderWrapper *yamlpatch.PlaceholderWrapper

	BeforeEach(func() {
		placeholderWrapper = yamlpatch.NewPlaceholderWrapper("{{", "}}")
	})

	Describe("Wrap", func() {
		It("returns the original content", func() {
			input := []byte(`content without any placeholders`)
			actual := placeholderWrapper.Wrap(input)
			Expect(actual).To(Equal(input))
		})

		It("returns the content with the placeholder wrapped when the content contains a placeholder", func() {
			input := []byte(`content with a {{placeholder}}`)
			expected := []byte(`content with a '{{placeholder}}'`)
			actual := placeholderWrapper.Wrap(input)
			Expect(actual).To(Equal(expected))
		})

		It("returns the content with the placeholder wrapped when the content contains a line with only a placeholder", func() {
			input := []byte(`
content: |
  {{placeholder}}
			`)
			expected := []byte(`
content: |
  '{{placeholder}}'
			`)
			actual := placeholderWrapper.Wrap(input)
			Expect(actual).To(Equal(expected))
		})

		It("returns the original content when the content contains an already-wrapped placeholder", func() {
			input := []byte(`content with a wrapped '{{placeholder}}'`)
			actual := placeholderWrapper.Wrap(input)
			Expect(string(actual)).To(Equal(string(input)))
		})

		It("supports alternate placeholders", func() {
			placeholderWrapper = yamlpatch.NewPlaceholderWrapper("((", "))")
			input := []byte(`content with an ((alternate-placeholder))`)
			expected := []byte(`content with an '((alternate-placeholder))'`)
			actual := placeholderWrapper.Wrap(input)
			Expect(actual).To(Equal(expected))
		})
	})

	Describe("Unwrap", func() {
		It("returns the original content", func() {
			input := []byte(`content without any placeholders`)
			actual := placeholderWrapper.Unwrap(input)
			Expect(string(actual)).To(Equal(string(input)))
		})

		It("returns the content with the placeholder unwrapped when the content contains a wrapped placeholder", func() {
			input := []byte(`content with a '{{placeholder}}'`)
			expected := []byte(`content with a {{placeholder}}`)
			actual := placeholderWrapper.Unwrap(input)
			Expect(string(actual)).To(Equal(string(expected)))
		})

		It("returns the content with the placeholder unwrapped when the content contains a line with only a wrapped placeholder", func() {
			input := []byte(`
content: |
  '{{placeholder}}'
			`)
			expected := []byte(`
content: |
  {{placeholder}}
			`)
			actual := placeholderWrapper.Unwrap(input)
			Expect(string(actual)).To(Equal(string(expected)))
		})

		It("supports alternate placeholders", func() {
			placeholderWrapper = yamlpatch.NewPlaceholderWrapper("((", "))")
			input := []byte(`content with an '((alternate-placeholder))'`)
			expected := []byte(`content with an ((alternate-placeholder))`)
			actual := placeholderWrapper.Unwrap(input)
			Expect(actual).To(Equal(expected))
		})
	})
})
