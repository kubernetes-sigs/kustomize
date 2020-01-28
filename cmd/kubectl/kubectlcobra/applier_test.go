// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kubectlcobra

import (
	"context"
	"github.com/stretchr/testify/assert"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdtesting "k8s.io/kubectl/pkg/cmd/testing"
	"testing"
)

// The applier is currently hard to test, as the dependencies on the ApplyOptions and
// the resolver are hard to stub out. As we work to better separate the different
// responsibilities of the apply functionality, we should also make it easier to test.
// This provides some basic tests for now.

func TestApplierWithUnknownFile(t *testing.T) {
	tf := cmdtesting.NewTestFactory()
	defer tf.Cleanup()
	iostreams, _, _, _ := genericclioptions.NewTestIOStreams()
	cmd := NewCmdApply("base", tf, iostreams)

	applier := newApplier(tf, iostreams)
	filenames := []string{"file.yaml"}
	applier.applyOptions.DeleteFlags.FileNameFlags.Filenames = &filenames

	err := applier.Initialize(cmd)
	assert.NoError(t, err)

	ch := applier.Run(context.TODO())

	var events []Event
	for msg := range ch {
		events = append(events, msg)
	}

	if !assert.Equal(t, 1, len(events)) {
		return
	}

	event := events[0]
	if !assert.Equal(t, ErrorEventType, event.EventType) {
		return
	}
	assert.Contains(t, event.ErrorEvent.Err.Error(), "does not exist")
}
