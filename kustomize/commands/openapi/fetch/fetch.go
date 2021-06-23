package fetch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"

	"github.com/spf13/cobra"
)

// NewCmdFetch makes a new fetch command.
func NewCmdFetch(w io.Writer) *cobra.Command {
	infoCmd := cobra.Command{
		Use: "fetch",
		Short: `Fetches the OpenAPI specification from the current kubernetes cluster specified 
in the user's kubeconfig`,
		Example: `kustomize openapi fetch`,
		Run: func(cmd *cobra.Command, args []string) {
			printSchema(w)
		},
	}
	return &infoCmd
}

func printSchema(w io.Writer) {
	errMsg := `
Error fetching schema from cluster.
Please make sure kubectl is installed and its context is set correctly.
Installation and setup instructions: https://kubernetes.io/docs/tasks/tools/install-kubectl/`

	command := exec.Command("kubectl", []string{"get", "--raw", "/openapi/v2"}...)
	var stderr bytes.Buffer
	var stdout bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	if err != nil || stdout.String() == "" {
		fmt.Fprintln(w, err, stderr.String()+errMsg)
		return
	}

	// format and output
	var jsonSchema map[string]interface{}
	output := stdout.Bytes()
	json.Unmarshal(output, &jsonSchema)
	output, _ = json.MarshalIndent(jsonSchema, "", "  ")
	fmt.Fprintln(w, string(output))
}
