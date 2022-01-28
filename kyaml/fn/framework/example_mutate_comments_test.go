package framework_test

import (
	"os"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
)

// In this example, we mutate line comments for field metadata.name.
// Some function may want to store some information in the comments (e.g.
// apply-setters function: https://catalog.kpt.dev/apply-setters/v0.2/)

func Example_dMutateComments() {
	if err := command.AsMain(framework.ResourceListProcessorFunc(mutateComments)); err != nil {
		os.Exit(1)
	}
}

func mutateComments(rl *framework.ResourceList) error {
	for i := range rl.Items {
		lineComment, err := rl.Items[i].GetLineComment("metadata", "name")
		if err != nil {
			return err
		}

		if strings.TrimSpace(lineComment) == "" {
			lineComment = "# bar-system"
		} else {
			lineComment = strings.Replace(lineComment, "foo", "bar", -1)
		}
		if err = rl.Items[i].SetLineComment(lineComment, "metadata", "name"); err != nil {
			return err
		}
	}
	return nil
}
