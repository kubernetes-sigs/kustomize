// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kubectlcobra

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestKubectlPrinterAdapter(t *testing.T) {
	ch := make(chan Event)
	buffer := bytes.Buffer{}
	operation := "operation"

	adapter := KubectlPrinterAdapter{
		ch: ch,
	}

	toPrinterFunc := adapter.toPrinterFunc()
	resourcePrinter, err := toPrinterFunc(operation)
	assert.NoError(t, err)

	deployment := appsv1.Deployment{
		TypeMeta: v1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "name",
			Namespace: "namespace",
		},
	}

	// Need to run this in a separate gorutine since go channels
	// are blocking.
	go func() {
		err = resourcePrinter.PrintObj(&deployment, &buffer)
	}()
	msg := <-ch

	assert.NoError(t, err)
	assert.Equal(t, operation, msg.ApplyEvent.Operation)
	assert.Equal(t, &deployment, msg.ApplyEvent.Object)
}
