package cmd

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kstatus/status"
	"sigs.k8s.io/kustomize/kstatus/wait"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

func GetFetchRunner() *FetchRunner {
	r := &FetchRunner{}
	c := &cobra.Command{
		Use:   "fetch",
		Short: "Fetch",
		RunE:  r.runE,
	}
	c.Flags().BoolVar(&r.IncludeSubpackages, "include-subpackages", true,
		"also print resources from subpackages.")

	r.Command = c
	return r
}

func FetchCommand() *cobra.Command {
	return GetFetchRunner().Command
}

type FetchRunner struct {
	IncludeSubpackages bool
	Command            *cobra.Command
}

func (r *FetchRunner) runE(c *cobra.Command, args []string) error {
	ctx := context.Background()
	client, err := getClient()
	if err != nil {
		return errors.Wrap(err, "error creating client")
	}

	resolver := wait.NewResolver(client, time.Minute)

	captureFilter := &CaptureIdentifiersFilter{}
	filters := []kio.Filter{captureFilter}

	var inputs []kio.Reader
	for _, a := range args {
		inputs = append(inputs, kio.LocalPackageReader{
			PackagePath:        a,
			IncludeSubpackages: r.IncludeSubpackages,
		})
	}
	if len(inputs) == 0 {
		inputs = append(inputs, &kio.ByteReader{Reader: c.InOrStdin()})
	}

	err = kio.Pipeline{
		Inputs:  inputs,
		Filters: filters,
	}.Execute()
	if err != nil {
		return errors.Wrap(err, "error reading manifests")
	}

	results := resolver.FetchAndResolve(ctx, captureFilter.Identifiers)

	newTablePrinter(FetchStatusInfo{results}, c.OutOrStdout(), c.OutOrStderr(), false).Print()
	return nil
}

type FetchStatusInfo struct {
	Results []wait.ResourceResult
}

func (f FetchStatusInfo) CurrentStatus() StatusData {
	var resourceData []ResourceStatusData
	for _, res := range f.Results {
		rsd := ResourceStatusData{
			Identifier: res.Resource,
		}
		if res.Error != nil {
			rsd.Status = status.UnknownStatus
			rsd.Message = res.Error.Error()
		} else {
			rsd.Status = res.Result.Status
			rsd.Message = res.Result.Message
		}

		resourceData = append(resourceData, rsd)
	}

	return StatusData{
		AggregateStatus:  status.UnknownStatus,
		ResourceStatuses: resourceData,
	}
}
