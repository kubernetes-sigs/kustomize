GO                ?= $(shell which go)
GIT_UPDATE_INDEX  := $(shell git update-index --refresh)
GIT_COMMIT        ?= $(shell git rev-parse HEAD)
GIT_VERSION       ?= $(shell git describe --tags --always --match "kustomize/*" | cut -d '/' -f2 2>/dev/null)

GOENV_GOOS               := $(shell go env GOOS)
GOENV_GOARCH             := $(shell go env GOARCH)
GOENV_GOARM              := $(shell go env GOARM)
GOOS                     ?= $(GOENV_GOOS)
GOARCH                   ?= $(GOENV_GOARCH)
GOARM                    ?= $(GOENV_GOARM)
GO_BUILD_SRC             := $(shell find . ../api -name \*.go -type f) go.mod go.sum
GO_BUILD_DIST            ?= ../dist
GO_BUILD_EXTLDFLAGS      :=
GO_BUILD_TAGS            := static
GO_BUILD_FLAGS           := -trimpath
GO_BUILD_LDFLAGS_OPTIMS  :=

ifeq ($(GOOS)/$(GOARCH),$(GOENV_GOOS)/$(GOENV_GOARCH))
GO_BUILD_TARGET          ?= $(GO_BUILD_DIST)/kustomize
GO_BUILD_VERSION_TARGET  ?= $(GO_BUILD_DIST)/kustomize-$(GIT_VERSION)
else
ifeq ($(GOARCH),arm)
GO_BUILD_TARGET          ?= $(GO_BUILD_DIST)/kustomize-$(GOOS)-$(GOARCH)v$(GOARM)
GO_BUILD_VERSION_TARGET  := $(GO_BUILD_DIST)/kustomize-$(GIT_VERSION)-$(GOOS)-$(GOARCH)v$(GOARM)
else
GO_BUILD_TARGET          ?= $(GO_BUILD_DIST)/kustomize-$(GOOS)-$(GOARCH)
GO_BUILD_VERSION_TARGET  := $(GO_BUILD_DIST)/kustomize-$(GIT_VERSION)-$(GOOS)-$(GOARCH)
endif # ($(GOARCH),arm)
endif # ($(GOOS)/$(GOARCH),$(GOENV_GOOS)/$(GOENV_GOARCH))

ifneq ($(DEBUG),)
GO_BUILD_FLAGS           += -race -gcflags="all=-N -l"
else
GO_BUILD_LDFLAGS_OPTIMS  += -s -w
endif # $(DEBUG)

GO_BUILD_FLAGS_TARGET           := .go-build-flags
GO_CROSSBUILD_WINDOWS_PLATFORMS := windows/amd64 windows/386 windows/arm
GO_CROSSBUILD_PLATFORMS         ?= linux/386 linux/amd64 linux/arm linux/arm64 linux/arm/v7 linux/arm/v6 \
                                   linux/mips linux/mips/softfloat linux/mips64 linux/mips64le linux/mipsle \
                                   linux/riscv64 linux/s390x \
                                   freebsd/386 freebsd/amd64 freebsd/arm freebsd/arm64 freebsd/arm/v7 freebsd/arm/v6 \
                                   openbsd/386 openbsd/amd64 openbsd/arm openbsd/arm64 openbsd/arm/v7 openbsd/arm/v6 \
                                   openbsd/mips64 \
                                   netbsd/386 netbsd/amd64 netbsd/arm netbsd/arm64 netbsd/arm/v7 netbsd/arm/v6 \
                                   plan9/386 plan9/amd64 plan9/arm plan9/arm/v7 plan9/arm/v6 \
                                   darwin/amd64 darwin/arm64 \
                                   dragonfly/amd64 illumos/amd64 solaris/amd64

GO_CROSSBUILD_386_PLATFORMS           := $(filter %/386,$(GO_CROSSBUILD_PLATFORMS))
GO_CROSSBUILD_AMD64_PLATFORMS         := $(filter %/amd64,$(GO_CROSSBUILD_PLATFORMS))
GO_CROSSBUILD_ARM_PLATFORMS           := $(filter %/arm,$(GO_CROSSBUILD_PLATFORMS))
GO_CROSSBUILD_ARM64_PLATFORMS         := $(filter %/arm64,$(GO_CROSSBUILD_PLATFORMS))
GO_CROSSBUILD_ARMV6_PLATFORMS         := $(filter %/arm/v6,$(GO_CROSSBUILD_PLATFORMS))
GO_CROSSBUILD_ARMV7_PLATFORMS         := $(filter %/arm/v7,$(GO_CROSSBUILD_PLATFORMS))
GO_CROSSBUILD_MIPS_PLATFORMS          := $(filter %/mips,$(GO_CROSSBUILD_PLATFORMS))
GO_CROSSBUILD_MIPSLE_PLATFORMS        := $(filter %/mipsle,$(GO_CROSSBUILD_PLATFORMS))
GO_CROSSBUILD_MIPS64_PLATFORMS        := $(filter %/mips64,$(GO_CROSSBUILD_PLATFORMS))
GO_CROSSBUILD_MIPS64LE_PLATFORMS      := $(filter %/mips64le,$(GO_CROSSBUILD_PLATFORMS))
GO_CROSSBUILD_MIPSSOFTFLOAT_PLATFORMS := $(filter %/mips/softfloat,$(GO_CROSSBUILD_PLATFORMS))
GO_CROSSBUILD_RISCV64_PLATFORMS       := $(filter %/riscv64,$(GO_CROSSBUILD_PLATFORMS))
GO_CROSSBUILD_S390X_PLATFORMS         := $(filter %/s390x,$(GO_CROSSBUILD_PLATFORMS))
GO_CROSSBUILD_WINDOWS_PLATFORMS       := $(filter windows/%,$(GO_CROSSBUILD_WINDOWS_PLATFORMS))

GO_CROSSBUILD_386_TARGET_PATTERN      := $(GO_BUILD_DIST)/kustomize-$(GIT_VERSION)-%-386
GO_CROSSBUILD_AMD64_TARGET_PATTERN    := $(GO_BUILD_DIST)/kustomize-$(GIT_VERSION)-%-amd64
GO_CROSSBUILD_ARM_TARGET_PATTERN      := $(GO_BUILD_DIST)/kustomize-$(GIT_VERSION)-%-arm
GO_CROSSBUILD_ARM64_TARGET_PATTERN    := $(GO_BUILD_DIST)/kustomize-$(GIT_VERSION)-%-arm64
GO_CROSSBUILD_ARMV6_TARGET_PATTERN    := $(GO_BUILD_DIST)/kustomize-$(GIT_VERSION)-%-armv6
GO_CROSSBUILD_ARMV7_TARGET_PATTERN    := $(GO_BUILD_DIST)/kustomize-$(GIT_VERSION)-%-armv7
GO_CROSSBUILD_MIPS_TARGET_PATTERN     := $(GO_BUILD_DIST)/kustomize-$(GIT_VERSION)-%-mips
GO_CROSSBUILD_MIPSLE_TARGET_PATTERN   := $(GO_BUILD_DIST)/kustomize-$(GIT_VERSION)-%-mipsle
GO_CROSSBUILD_MIPS64_TARGET_PATTERN   := $(GO_BUILD_DIST)/kustomize-$(GIT_VERSION)-%-mips64
GO_CROSSBUILD_MIPS64LE_TARGET_PATTERN := $(GO_BUILD_DIST)/kustomize-$(GIT_VERSION)-%-mips64le
GO_CROSSBUILD_RISCV64_TARGET_PATTERN  := $(GO_BUILD_DIST)/kustomize-$(GIT_VERSION)-%-riscv64
GO_CROSSBUILD_S390X_TARGET_PATTERN    := $(GO_BUILD_DIST)/kustomize-$(GIT_VERSION)-%-s390x
GO_CROSSBUILD_WINDOWS_TARGET_PATTERN  := $(GO_BUILD_DIST)/kustomize-$(GIT_VERSION)-windows-%.exe

GO_CROSSBUILD_TARGETS := $(patsubst %/386,$(GO_CROSSBUILD_386_TARGET_PATTERN),$(GO_CROSSBUILD_386_PLATFORMS))
GO_CROSSBUILD_TARGETS += $(patsubst %/amd64,$(GO_CROSSBUILD_AMD64_TARGET_PATTERN),$(GO_CROSSBUILD_AMD64_PLATFORMS))
GO_CROSSBUILD_TARGETS += $(patsubst %/arm,$(GO_CROSSBUILD_ARM_TARGET_PATTERN),$(GO_CROSSBUILD_ARM_PLATFORMS))
GO_CROSSBUILD_TARGETS += $(patsubst %/arm64,$(GO_CROSSBUILD_ARM64_TARGET_PATTERN),$(GO_CROSSBUILD_ARM64_PLATFORMS))
GO_CROSSBUILD_TARGETS += $(patsubst %/arm/v6,$(GO_CROSSBUILD_ARMV6_TARGET_PATTERN),$(GO_CROSSBUILD_ARMV6_PLATFORMS))
GO_CROSSBUILD_TARGETS += $(patsubst %/arm/v7,$(GO_CROSSBUILD_ARMV7_TARGET_PATTERN),$(GO_CROSSBUILD_ARMV7_PLATFORMS))
GO_CROSSBUILD_TARGETS += $(patsubst %/mips,$(GO_CROSSBUILD_MIPS_TARGET_PATTERN),$(GO_CROSSBUILD_MIPS_PLATFORMS))
GO_CROSSBUILD_TARGETS += $(patsubst %/mipsle,$(GO_CROSSBUILD_MIPSLE_TARGET_PATTERN),$(GO_CROSSBUILD_MIPSLE_PLATFORMS))
GO_CROSSBUILD_TARGETS += $(patsubst %/mips64,$(GO_CROSSBUILD_MIPS64_TARGET_PATTERN),$(GO_CROSSBUILD_MIPS64_PLATFORMS))
GO_CROSSBUILD_TARGETS += $(patsubst %/mips64le,$(GO_CROSSBUILD_MIPS64LE_TARGET_PATTERN),$(GO_CROSSBUILD_MIPS64LE_PLATFORMS))
GO_CROSSBUILD_TARGETS += $(patsubst %/mips/softfloat,$(GO_CROSSBUILD_LINUX_MIPS_TARGET_PATTERN),$(GO_CROSSBUILD_LINUX_MIPS_PLATFORMS))
GO_CROSSBUILD_TARGETS += $(patsubst %/riscv64,$(GO_CROSSBUILD_RISCV64_TARGET_PATTERN),$(GO_CROSSBUILD_RISCV64_PLATFORMS))
GO_CROSSBUILD_TARGETS += $(patsubst %/s390x,$(GO_CROSSBUILD_S390X_TARGET_PATTERN),$(GO_CROSSBUILD_S390X_PLATFORMS))
GO_CROSSBUILD_TARGETS += $(patsubst windows/%,$(GO_CROSSBUILD_WINDOWS_TARGET_PATTERN),$(GO_CROSSBUILD_WINDOWS_PLATFORMS))

GO_BUILD_EXTLDFLAGS     := $(strip $(GO_BUILD_EXTLDFLAGS))
GO_BUILD_TAGS           := $(strip $(GO_BUILD_TAGS))
GO_BUILD_FLAGS          := $(strip $(GO_BUILD_FLAGS))
GO_BUILD_LDFLAGS_OPTIMS := $(strip $(GO_BUILD_LDFLAGS_OPTIMS))
GO_BUILD_LDFLAGS        := -ldflags '$(GO_BUILD_LDFLAGS_OPTIMS) -X sigs.k8s.io/kustomize/api/provenance.version=$(GIT_VERSION) \
-X sigs.k8s.io/kustomize/api/provenance.gitCommit=$(GIT_COMMIT) \
-X sigs.k8s.io/kustomize/api/provenance.buildDate=$(shell date -u +%FT%TZ) -extldflags "$(GO_BUILD_EXTLDFLAGS)"'

# ------------------------------------------------------------------------------

.PHONY: all build crossbuild crossbuild-checksums .FORCE

all: crossbuild crossbuild-checksums

build: $(GO_BUILD_VERSION_TARGET) $(GO_BUILD_TARGET)

install:
	CGO_ENABLED=0 $(GO) install -tags $(GO_BUILD_TAGS) $(GO_BUILD_FLAGS) $(GO_BUILD_LDFLAGS)

$(GO_BUILD_FLAGS_TARGET) : .FORCE
	@(echo "GO_VERSION=$(shell $(GO) version)"; \
	  echo "GO_GOOS=$(GOOS)"; \
	  echo "GO_GOARCH=$(GOARCH)"; \
	  echo "GO_GOARM=$(GOARM)"; \
	  echo "GO_BUILD_TAGS=$(GO_BUILD_TAGS)"; \
	  echo "GO_BUILD_FLAGS=$(GO_BUILD_FLAGS)"; \
	  echo "GO_BUILD_LDFLAGS_OPTIMS=$(GO_BUILD_LDFLAGS_OPTIMS)") \
	    | cmp -s - $@ \
	        || (echo "GO_VERSION=$(shell $(GO) version)"; \
	            echo "GO_GOOS=$(GOOS)"; \
	            echo "GO_GOARCH=$(GOARCH)"; \
	            echo "GO_GOARM=$(GOARM)"; \
	            echo "GO_BUILD_TAGS=$(GO_BUILD_TAGS)"; \
	            echo "GO_BUILD_FLAGS=$(GO_BUILD_FLAGS)"; \
	            echo "GO_BUILD_LDFLAGS_OPTIMS=$(GO_BUILD_LDFLAGS_OPTIMS)") > $@

$(GO_BUILD_TARGET): $(GO_BUILD_VERSION_TARGET)
	@(test -e $@ && unlink $@) || true
	@mkdir -p $$(dirname $@)
	@ln $< $@

$(GO_BUILD_VERSION_TARGET): $(GO_BUILD_SRC) $(GO_GENERATE_TARGET) $(GO_BUILD_FLAGS_TARGET)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM) $(GO) build -tags $(GO_BUILD_TAGS) $(GO_BUILD_FLAGS) $(GO_BUILD_LDFLAGS) -o $@

crossbuild: $(GO_BUILD_VERSION_TARGET) $(GO_CROSSBUILD_TARGETS) crossbuild-checksums

$(GO_CROSSBUILD_WINDOWS_TARGET_PATTERN): $(GO_BUILD_SRC) $(GO_BUILD_FLAGS_TARGET)
	CGO_ENABLED=0 GOOS=windows GOARCH=$* $(GO) build -tags $(GO_BUILD_TAGS),crossbuild $(GO_BUILD_FLAGS) $(GO_BUILD_LDFLAGS) -o $@

$(GO_CROSSBUILD_386_TARGET_PATTERN): $(GO_BUILD_SRC) $(GO_BUILD_FLAGS_TARGET)
	CGO_ENABLED=0 GOOS=$* GOARCH=386 $(GO) build -tags $(GO_BUILD_TAGS),crossbuild $(GO_BUILD_FLAGS) $(GO_BUILD_LDFLAGS) -o $@

$(GO_CROSSBUILD_AMD64_TARGET_PATTERN): $(GO_BUILD_SRC) $(GO_BUILD_FLAGS_TARGET)
	CGO_ENABLED=0 GOOS=$* GOARCH=amd64 $(GO) build -tags $(GO_BUILD_TAGS),crossbuild $(GO_BUILD_FLAGS) $(GO_BUILD_LDFLAGS) -o $@

$(GO_CROSSBUILD_ARM_TARGET_PATTERN): $(GO_BUILD_SRC) $(GO_BUILD_FLAGS_TARGET)
	CGO_ENABLED=0 GOOS=$* GOARCH=arm $(GO) build -tags $(GO_BUILD_TAGS),crossbuild $(GO_BUILD_FLAGS) $(GO_BUILD_LDFLAGS) -o $@

$(GO_CROSSBUILD_ARM64_TARGET_PATTERN): $(GO_BUILD_SRC) $(GO_BUILD_FLAGS_TARGET)
	CGO_ENABLED=0 GOOS=$* GOARCH=arm64 $(GO) build -tags $(GO_BUILD_TAGS),crossbuild $(GO_BUILD_FLAGS) $(GO_BUILD_LDFLAGS) -o $@

$(GO_CROSSBUILD_ARMV6_TARGET_PATTERN): $(GO_BUILD_SRC) $(GO_BUILD_FLAGS_TARGET)
	CGO_ENABLED=0 GOOS=$* GOARCH=arm GOARM=6 $(GO) build -tags $(GO_BUILD_TAGS),crossbuild $(GO_BUILD_FLAGS) $(GO_BUILD_LDFLAGS) -o $@

$(GO_CROSSBUILD_ARMV7_TARGET_PATTERN): $(GO_BUILD_SRC) $(GO_BUILD_FLAGS_TARGET)
	CGO_ENABLED=0 GOOS=$* GOARCH=arm GOARM=7 $(GO) build -tags $(GO_BUILD_TAGS),crossbuild $(GO_BUILD_FLAGS) $(GO_BUILD_LDFLAGS) -o $@

$(GO_CROSSBUILD_MIPS_TARGET_PATTERN): $(GO_BUILD_SRC) $(GO_BUILD_FLAGS_TARGET)
	CGO_ENABLED=0 GOOS=$* GOARCH=mips $(GO) build -tags $(GO_BUILD_TAGS),crossbuild $(GO_BUILD_FLAGS) $(GO_BUILD_LDFLAGS) -o $@

$(GO_CROSSBUILD_MIPSLE_TARGET_PATTERN): $(GO_BUILD_SRC) $(GO_BUILD_FLAGS_TARGET)
	CGO_ENABLED=0 GOOS=$* GOARCH=mipsle $(GO) build -tags $(GO_BUILD_TAGS),crossbuild $(GO_BUILD_FLAGS) $(GO_BUILD_LDFLAGS) -o $@

$(GO_CROSSBUILD_MIPS64_TARGET_PATTERN): $(GO_BUILD_SRC) $(GO_BUILD_FLAGS_TARGET)
	CGO_ENABLED=0 GOOS=$* GOARCH=mips64 $(GO) build -tags $(GO_BUILD_TAGS),crossbuild $(GO_BUILD_FLAGS) $(GO_BUILD_LDFLAGS) -o $@

$(GO_CROSSBUILD_MIPS64LE_TARGET_PATTERN): $(GO_BUILD_SRC) $(GO_BUILD_FLAGS_TARGET)
	CGO_ENABLED=0 GOOS=$* GOARCH=mips64le $(GO) build -tags $(GO_BUILD_TAGS),crossbuild $(GO_BUILD_FLAGS) $(GO_BUILD_LDFLAGS) -o $@

$(GO_CROSSBUILD_MIPSSOFFLOAT_TARGET_PATTERN): $(GO_BUILD_SRC) $(GO_BUILD_FLAGS_TARGET)
	CGO_ENABLED=0 GOOS=$* GOARCH=mips GOMIPS=softfloat $(GO) build -tags $(GO_BUILD_TAGS),crossbuild $(GO_BUILD_FLAGS) $(GO_BUILD_LDFLAGS) -o $@

$(GO_CROSSBUILD_RISCV64_TARGET_PATTERN): $(GO_BUILD_SRC) $(GO_BUILD_FLAGS_TARGET)
	CGO_ENABLED=0 GOOS=$* GOARCH=riscv64 $(GO) build -tags $(GO_BUILD_TAGS),crossbuild $(GO_BUILD_FLAGS) $(GO_BUILD_LDFLAGS) -o $@

$(GO_CROSSBUILD_S390X_TARGET_PATTERN): $(GO_BUILD_SRC) $(GO_BUILD_FLAGS_TARGET)
	CGO_ENABLED=0 GOOS=$* GOARCH=s390x $(GO) build -tags $(GO_BUILD_TAGS),crossbuild $(GO_BUILD_FLAGS) $(GO_BUILD_LDFLAGS) -o $@

crossbuild-checksums: $(GO_BUILD_DIST)/checksums

$(GO_BUILD_DIST)/checksums : $(GO_CROSSBUILD_TARGETS)
	cd $(GO_BUILD_DIST) && shasum -a 256 kustomize-*-* > checksums
