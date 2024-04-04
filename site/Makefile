# Credit to the Kubernetes Website team. (https://github.com/kubernetes/website/blob/main/Makefile)
HUGO_VERSION      = $(shell grep ^HUGO_VERSION netlify.toml | tail -n 1 | cut -d '=' -f 2 | tr -d " \"\n")

# The CONTAINER_ENGINE variable is used for specifying the container engine. By default 'docker' is used
# but this can be overridden when calling make, e.g.
# CONTAINER_ENGINE=podman make container-image
CONTAINER_ENGINE ?= docker
IMAGE_REGISTRY ?= registry.local:5000
IMAGE_VERSION=$(shell scripts/hash-files.sh Dockerfile Makefile | cut -c 1-12)
CONTAINER_IMAGE   = $(IMAGE_REGISTRY)/kustomize-website-hugo:v$(HUGO_VERSION)-$(IMAGE_VERSION)
# Mount read-only to allow use with tools like Podman in SELinux mode
# Container targets don't need to write into /src
CONTAINER_RUN     = "$(CONTAINER_ENGINE)" run --rm --interactive --tty --volume "$(abspath $(CURDIR)/..):/src:ro,Z"

CCRED=\033[0;31m
CCEND=\033[0m

.PHONY: help
help: ## Show this help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: module-check
module-check: ## Check if all of the required submodules are correctly initialized.
	@git submodule status --recursive | awk '/^[+-]/ {err = 1; printf "\033[31mWARNING\033[0m Submodule not initialized: \033[34m%s\033[0m\n",$$2} END { if (err != 0) print "You need to run \033[32mmake module-init\033[0m to initialize missing modules first"; exit err }' 1>&2

.PHONY: module-init
module-init: ## Initialize required submodules.
	@echo "Initializing submodules..." 1>&2
	@git submodule update --init --recursive --depth 1

.PHONY: all
all: build ## Build site with production settings and put deliverables in ./public

.PHONY: build
build: module-check ## Build site with non-production settings and put deliverables in ./public
	hugo --cleanDestinationDir --minify --environment development

.PHONY: build-preview
build-preview: module-check ## Build site with drafts and future posts enabled
	hugo --cleanDestinationDir --buildDrafts --buildFuture --environment preview

.PHONY: serve
serve: module-check ## Boot the development server.
	hugo server --buildFuture --environment development

## Update the hugo version in ../hack/go.mod to match HUGO_VERSION in netlify.toml
.PHONY: tools
tools: netlify.toml
	@echo "Hugo: v$(HUGO_VERSION)"
	@set -euo pipefail && cd ../hack && \
		sed -e "/github.com\/gohugoio\/hugo/ s/v[0-9]\+\.[0-9]\+\.[0-9]\+/v$(HUGO_VERSION)/" go.mod > go.mod.next && \
		mv go.mod.next go.mod && \
		echo "Updated ../hack/go.mod"
	cd ../hack && go mod tidy

## Build a container image for the preview of the website
.PHONY: container-image
container-image: tools Dockerfile hugo.toml
	cd .. && $(CONTAINER_ENGINE) build . \
		--file site/Dockerfile \
		--network=host \
		--tag $(CONTAINER_IMAGE)

# no build lock to allow for read-only mounts
## Boot the development server using container.
.PHONY: container-serve
container-serve: module-check
	$(CONTAINER_RUN) \
		--cap-drop=ALL \
		--cap-add=AUDIT_WRITE \
		--read-only \
		--mount type=tmpfs,destination=/tmp,tmpfs-mode=01777 \
		-p 1313:1313 $(CONTAINER_IMAGE) \
		hugo server \
		  --source site \
		  --buildFuture \
		  --environment development \
		  --bind 0.0.0.0 \
		  --destination /tmp/hugo \
		  --cleanDestinationDir \
		  --noBuildLock
