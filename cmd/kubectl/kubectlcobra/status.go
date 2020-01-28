// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kubectlcobra

import (
	"time"

	"github.com/spf13/cobra"
)

func NewStatusOptions() *StatusOptions {
	return &StatusOptions{
		wait:    false,
		period:  2 * time.Second,
		timeout: time.Minute,
	}
}

type StatusOptions struct {
	wait    bool
	period  time.Duration
	timeout time.Duration
}

func (s *StatusOptions) AddFlags(c *cobra.Command) {
	c.Flags().BoolVar(&s.wait, "status", s.wait, "Wait for all applied resources to reach the Current status.")
	c.Flags().DurationVar(&s.period, "status-period", s.period, "Polling period for resource statuses.")
	c.Flags().DurationVar(&s.timeout, "status-timeout", s.timeout, "Timeout threshold for waiting for all resources to reach the Current status.")
}
