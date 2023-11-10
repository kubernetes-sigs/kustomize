// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package utils_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	. "sigs.k8s.io/kustomize/api/internal/utils"
)

const (
	timeToWait = 10 * time.Millisecond
	tooSlow    = 2 * timeToWait
)

func errMsg(msg string) string {
	return fmt.Sprintf("hit %s timeout running '%s'", timeToWait, msg)
}

func TestTimedCallFastNoError(t *testing.T) {
	err := TimedCall(
		"fast no error", timeToWait,
		func() error { return nil })
	if !assert.NoError(t, err) {
		t.Fatal(err)
	}
}

func TestTimedCallFastWithError(t *testing.T) {
	err := TimedCall(
		"fast with error", timeToWait,
		func() error { return assert.AnError })
	if assert.Error(t, err) {
		assert.EqualError(t, err, assert.AnError.Error())
	} else {
		t.Fail()
	}
}

func TestTimedCallSlowNoError(t *testing.T) {
	err := TimedCall(
		"slow no error", timeToWait,
		func() error { time.Sleep(tooSlow); return nil })
	if assert.Error(t, err) {
		assert.EqualError(t, err, errMsg("slow no error"))
	} else {
		t.Fail()
	}
}

func TestTimedCallSlowWithError(t *testing.T) {
	err := TimedCall(
		"slow with error", timeToWait,
		func() error { time.Sleep(tooSlow); return assert.AnError })
	if assert.Error(t, err) {
		assert.EqualError(t, err, errMsg("slow with error"))
	} else {
		t.Fail()
	}
}

func TestTimedCallGoroutineLeak(t *testing.T) {
	defer goleak.VerifyNone(t)
	err := TimedCall("function done, no goroutine leaks", timeToWait, func() error {
		time.Sleep(tooSlow)
		return fmt.Errorf("function done")
	})
	if assert.Error(t, err) {
		assert.EqualError(t, err, errMsg("function done, no goroutine leaks"))
	} else {
		t.Fail()
	}

	// The code introduces a 2-second sleep to allow the goroutine to complete its execution.
	// Subsequently, it verifies if the goroutine created by the function exits as expected.
	time.Sleep(tooSlow)
}
