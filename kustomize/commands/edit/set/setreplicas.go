// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package set

import (
	"errors"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/internal/kustfile"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type setReplicasOptions struct {
	replicasMap map[string]types.Replica
}

// errors

var (
	errReplicasNoArgs      = errors.New("no replicas specified")
	errReplicasInvalidArgs = errors.New(`invalid format of replica, use the following format: <name>=<count>`)
)

const replicasSeparator = "="

// newCmdSetReplicas sets the new replica count for a resource in the kustomization.
func newCmdSetReplicas(fSys filesys.FileSystem) *cobra.Command {
	var o setReplicasOptions

	cmd := &cobra.Command{
		Use:   "replicas",
		Short: `Sets replicas count for resources in the kustomization file`,
		Example: `
The command
  set replicas my-app=3 other-app=1
will add

replicas:
- name: my-app
  count: 3
- name: other-app
  count: 1

to the kustomization file if it doesn't exist,
and overwrite the previous ones if the replicas name exists.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate(args)
			if err != nil {
				return err
			}
			return o.RunSetReplicas(fSys)
		},
	}
	return cmd
}

// Validate validates setImage command.
func (o *setReplicasOptions) Validate(args []string) error {
	if len(args) == 0 {
		return errReplicasNoArgs
	}

	o.replicasMap = make(map[string]types.Replica)

	for _, arg := range args {

		replica, err := parseReplicasArg(arg)
		if err != nil {
			return err
		}
		o.replicasMap[replica.Name] = replica
	}
	return nil
}

// RunSetReplicas runs setReplicas command.
func (o *setReplicasOptions) RunSetReplicas(fSys filesys.FileSystem) error {
	mf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		return err
	}
	m, err := mf.Read()
	if err != nil {
		return err
	}

	// append only new replicas from kustomize file
	for _, rep := range m.Replicas {
		if _, ok := o.replicasMap[rep.Name]; ok {
			continue
		}

		o.replicasMap[rep.Name] = rep
	}

	var replicas []types.Replica
	for _, v := range o.replicasMap {
		replicas = append(replicas, v)
	}

	sort.Slice(replicas, func(i, j int) bool {
		return replicas[i].Name < replicas[j].Name
	})

	m.Replicas = replicas
	return mf.Write(m)
}

func parseReplicasArg(arg string) (types.Replica, error) {

	// matches a name and a replica count
	// <name>=<count>
	if s := strings.Split(arg, replicasSeparator); len(s) == 2 {
		count, err := strconv.ParseInt(s[1], 10, 64)
		if err != nil {
			return types.Replica{}, errReplicasInvalidArgs
		}

		return types.Replica{
			Name:  s[0],
			Count: count,
		}, nil
	}

	return types.Replica{}, errReplicasInvalidArgs
}
