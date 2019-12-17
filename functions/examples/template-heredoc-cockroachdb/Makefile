# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

.PHONY: license image

license:
	(which $(GOPATH)/bin/addlicense || go get github.com/google/addlicense)
	$(GOPATH)/bin/addlicense  -y 2019 -c "The Kubernetes Authors." -f LICENSE_TEMPLATE .

all: license

image:
	docker build image -t gcr.io/kustomize-functions/example-cockroachdb:v0.1.0
	docker push gcr.io/kustomize-functions/example-cockroachdb:v0.1.0
