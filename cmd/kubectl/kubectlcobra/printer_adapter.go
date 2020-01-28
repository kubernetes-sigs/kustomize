// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kubectlcobra

import (
	"io"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/printers"
)

// KubectlPrinterAdapter is a workaround for capturing progress from
// ApplyOptions. ApplyOptions were originally meant to print progress
// directly using a configurable printer. The KubectlPrinterAdapter
// plugs into ApplyOptions as a ToPrinter function, but instead of
// printing the info, it emits it as an event on the provided channel.
type KubectlPrinterAdapter struct {
	ch chan<- Event
}

// resourcePrinterImpl implements the ResourcePrinter interface. But
// instead of printing, it emits information on the provided channel.
type resourcePrinterImpl struct {
	operation string
	ch        chan<- Event
}

// PrintObj takes the provided object and operation and emits
// it on the channel.
func (r *resourcePrinterImpl) PrintObj(obj runtime.Object, _ io.Writer) error {
	r.ch <- Event{
		EventType: ApplyEventType,
		ApplyEvent: ApplyEvent{
			Operation: r.operation,
			Object:    obj,
		},
	}
	return nil
}

type toPrinterFunc func(string) (printers.ResourcePrinter, error)

// toPrinterFunc returns a function of type toPrinterFunc. This
// is the type required by the ApplyOptions.
func (p *KubectlPrinterAdapter) toPrinterFunc() toPrinterFunc {
	return func(operation string) (printers.ResourcePrinter, error) {
		return &resourcePrinterImpl{
			ch:        p.ch,
			operation: operation,
		}, nil
	}
}
