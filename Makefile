# This Makefile is (and must be) used by
# travis/pre-commit.sh to qualify pull requests.
#
# That script generates all the code that needs
# to be generated, and runs all the tests.
#
# Functionality in that script, expressed in bash, is
# gradually being moved here.

MYGOBIN := $(shell go env GOPATH)/bin

.PHONY: all
all:
	./travis/pre-commit.sh

$(MYGOBIN)/golangci-lint:
	cd api; go install github.com/golangci/golangci-lint/cmd/golangci-lint

$(MYGOBIN)/mdrip:
	cd api; go install github.com/monopole/mdrip

# TODO: need a new release of the API, followed by a new pluginator.
# pluginator v1.1.0 is too old for the code currently needed in the API.
# Can release a new one at any time, just haven't done so.
# When one has been released,
#  - uncomment the pluginator line in './api/internal/tools/tools.go'
#  - pin the version tag in './api/go.mod' to match the new release
#  - change the following to 'cd api; go install sigs.k8s.io/kustomize/pluginator'
$(MYGOBIN)/pluginator:
	cd pluginator; go install .

$(MYGOBIN)/stringer:
	cd api; go install golang.org/x/tools/cmd/stringer

# Specific version tags for these utilities are pinned in ./api/go.mod
# which seems to be as good a place as any to do so.
# That's the reason for all the occurances of 'cd api;' in the
# dependencies; 'go install' uses the local 'go.mod' to get the version.
install-tools: $(MYGOBIN)/golangci-lint \
	$(MYGOBIN)/mdrip \
	$(MYGOBIN)/pluginator \
	$(MYGOBIN)/stringer

.PHONY: lint
lint: install-tools
	cd api; $(MYGOBIN)/golangci-lint run ./...
	cd kustomize; $(MYGOBIN)/golangci-lint run ./...
	cd pluginator; $(MYGOBIN)/golangci-lint run ./...

.PHONY: unit-test-api
unit-test-api:
	cd api; go test ./...

.PHONY: unit-test-plugins
unit-test-plugins:
	./hack/runPluginUnitTests.sh

.PHONY: unit-test-kustomize
unit-test-kustomize:
	cd kustomize; go test ./...

.PHONY: unit-test-all
unit-test-all: unit-test-api unit-test-kustomize unit-test-plugins

# linux only.
$(MYGOBIN)/kubeval:
	d=$(shell mktemp -d); cd $$d; \
	wget https://github.com/instrumenta/kubeval/releases/latest/download/kubeval-linux-amd64.tar.gz; \
	tar xf kubeval-linux-amd64.tar.gz; \
	mv kubeval $(MYGOBIN); \
	rm -rf $$d

# linux only.
$(MYGOBIN)/helm:
	d=$(shell mktemp -d); cd $$d; \
	wget https://storage.googleapis.com/kubernetes-helm/helm-v2.13.1-linux-amd64.tar.gz; \
	tar -xvzf helm-v2.13.1-linux-amd64.tar.gz; \
	mv linux-amd64/helm $(MYGOBIN); \
	rm -rf $$d
