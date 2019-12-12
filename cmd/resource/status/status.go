package status

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/resource/status/cmd"
)

func StatusCommand() *cobra.Command {
	var status = &cobra.Command{
		Use:   "status",
		Short: "status reference command",
	}

	status.AddCommand(cmd.FetchCommand())
	status.AddCommand(cmd.WaitCommand())
	status.AddCommand(cmd.EventsCommand())

	return status
}
