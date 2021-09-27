#!/usr/bin/env bash
# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

# TODO: Make the code ignorant of the CI environment "brand name".
# We used to run CI tests on travis, and disabled certain tests
# when running there.  Now we run on Prow, so look for that.
# https://github.com/kubernetes/test-infra/blob/master/prow/jobs.md
# Might be useful to eschew using the brand name of the CI environment
# (replace "travis" with "RemoteCI" or something - not just switch to "prow").

function onLinuxAndNotOnRemoteCI {
  [[ ("linux" == "$(go env GOOS)") && (-z ${PROW_JOB_ID+x}) ]] && return
  false
}
