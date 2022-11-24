.PHONY: test all clean generate run

FUNC_WRAPPER_SRC_DIR = funcwrappersrc
FUNC_WRAPPER_DST_DIR = funcwrapper

include ../../Makefile-modules.mk

all: test build

test:
	go test -v ./...

build:
	go build -o $(MYGOBIN) main.go

install:
	go install .

run:
	go run . $(ARGS)

clean:
	rm -f pluginator
