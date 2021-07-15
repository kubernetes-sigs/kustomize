package fetch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

var format string

// NewCmdFetch makes a new fetch command.
func NewCmdFetch(w io.Writer) *cobra.Command {
	fetchCmd := cobra.Command{
		Use: "fetch",
		Short: `Fetches the OpenAPI specification from the current kubernetes cluster specified 
in the user's kubeconfig`,
		Example: `kustomize openapi fetch`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return printSchema(w)
		},
	}
	fetchCmd.Flags().StringVar(
		&format,
		"format",
		"json",
		"Specify format for fetched schema ('json' or 'yaml')")
	return &fetchCmd
}

func printSchema(w io.Writer) error {
	if format != "json" && format != "yaml" {
		return fmt.Errorf("format must be either 'json' or 'yaml'")
	}

	errMsg := `
Error fetching schema from cluster.
Please make sure kubectl is installed, its context is set correctly, and your cluster is up.
Installation and setup instructions: https://kubernetes.io/docs/tasks/tools/install-kubectl/`

	command := exec.Command("kubectl", []string{"get", "--raw", "/openapi/v2"}...)
	var stderr bytes.Buffer
	var stdout bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	if err != nil {
		return fmt.Errorf("%w\n%s", err, stderr.String()+errMsg)
	} else if stdout.String() == "" {
		return fmt.Errorf(stderr.String() + errMsg)
	}

	// format and output
	var jsonSchema map[string]interface{}
	output := stdout.Bytes()
	json.Unmarshal(output, &jsonSchema)
	output, _ = json.MarshalIndent(jsonSchema, "", "  ")

	if format == "yaml" {
		output, err = yaml.JSONToYAML(output)
		if err != nil {
			return err
		}
	}

	fmt.Fprintln(w, string(output))
	return nil
}
