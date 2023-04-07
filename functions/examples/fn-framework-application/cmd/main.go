package main

import "sigs.k8s.io/kustomize/functions/examples/fn-framework-application/pkg/dispatcher"

func main() {
	_ = dispatcher.NewCommand().Execute()
}
