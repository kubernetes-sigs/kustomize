package main

import (
	"os"

	"github.com/spf13/cobra"

	"sigs.k8s.io/kustomize/cmd/resource/status"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var root = &cobra.Command{
	Use:   "resource",
	Short: "resource reference command",
}

func main() {
	root.AddCommand(status.StatusCommand())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
