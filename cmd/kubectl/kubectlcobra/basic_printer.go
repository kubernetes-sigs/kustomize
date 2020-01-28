// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kubectlcobra

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/kustomize/kstatus/wait"
)

// BasicPrinter is a simple implementation that just prints the events
// from the channel in the default format for kubectl.
// We need to support different printers for different output formats.
type BasicPrinter struct {
	ioStreams genericclioptions.IOStreams
}

// Print outputs the events from the provided channel in a simple
// format on StdOut. As we support other printer implementations
// this should probably be an interface.
// This function will block until the channel is closed.
func (b *BasicPrinter) Print(ch <-chan Event) {
	for event := range ch {
		switch event.EventType {
		case ErrorEventType:
			cmdutil.CheckErr(event.ErrorEvent.Err)
		case ApplyEventType:
			obj := event.ApplyEvent.Object
			gvk := obj.GetObjectKind().GroupVersionKind()
			name := "<unknown>"
			if acc, err := meta.Accessor(obj); err == nil {
				if n := acc.GetName(); len(n) > 0 {
					name = n
				}
			}
			fmt.Fprintf(b.ioStreams.Out, "%s %s\n", resourceIdToString(gvk.GroupKind(), name), event.ApplyEvent.Operation)
		case StatusEventType:
			statusEvent := event.StatusEvent
			switch statusEvent.Type {
			case wait.ResourceUpdate:
				id := statusEvent.EventResource.ResourceIdentifier
				gk := id.GroupKind
				fmt.Fprintf(b.ioStreams.Out, "%s is %s: %s\n", resourceIdToString(gk, id.Name), statusEvent.EventResource.Status.String(), statusEvent.EventResource.Message)
			case wait.Completed:
				fmt.Fprint(b.ioStreams.Out, "all resources has reached the Current status\n")
			case wait.Aborted:
				fmt.Fprintf(b.ioStreams.Out, "resources failed to the reached Current status\n")
			}
		}
	}
}

// resourceIdToString returns the string representation of a GroupKind and a resource name.
func resourceIdToString(gk schema.GroupKind, name string) string {
	return fmt.Sprintf("%s/%s", strings.ToLower(gk.String()), name)
}
