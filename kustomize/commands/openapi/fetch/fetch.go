package fetch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"time"

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
		Hidden: true,
	}

	return &infoCmd
}

func printSchema(w io.Writer) {
	fmt.Fprintln(w, "Fetching schema from cluster")
	errMsg := `
Error fetching schema from cluster.
Please make sure port 8081 is available, kubectl is installed, and its context is set correctly.
Installation and setup instructions: https://kubernetes.io/docs/tasks/tools/install-kubectl/`

	command := exec.Command("kubectl", []string{"proxy", "--port=8081", "&"}...)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	err := command.Start()
	defer killProcess(command)

	// give the proxy a second to start up
	time.Sleep(time.Second)

	if err != nil || stderr.String() != "" {
		fmt.Fprintln(w, err, stderr.String()+errMsg)
		return
	}

	commandCurl := exec.Command("curl", []string{"http://localhost:8081/openapi/v2"}...)
	var stdout bytes.Buffer
	commandCurl.Stdout = &stdout
	commandCurl.Stderr = &stderr
	err = commandCurl.Run()
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

func killProcess(command *exec.Cmd) {
	if command.Process != nil {
		command.Process.Kill()
	}
}
