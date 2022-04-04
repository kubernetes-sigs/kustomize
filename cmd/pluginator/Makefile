.PHONY: statik test all clean generate run

FUNC_WRAPPER_SRC_DIR = funcwrappersrc
FUNC_WRAPPER_DST_DIR = funcwrapper

include ../../Makefile-modules.mk

all: test build

test: generate
	go test ./...

generate: $(MYGOBIN)/statik
	( \
		cd internal/krmfunction; \
		$(MYGOBIN)/statik -src=$(FUNC_WRAPPER_SRC_DIR) -f -p $(FUNC_WRAPPER_DST_DIR) -include=main.go,go.mod.src \
	)

build: generate
	go build -o pluginator main.go

install: generate
	go install .

run: generate
	go run . $(ARGS)

clean:
	rm -f pluginator
