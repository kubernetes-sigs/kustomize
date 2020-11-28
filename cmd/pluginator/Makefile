.PHONY: statik test all clean generate run

FUNC_WRAPPER_SRC_DIR = funcwrappersrc
FUNC_WRAPPER_DST_DIR = funcwrapper

all: test build

statik:
	go get github.com/rakyll/statik

test: generate
	go test ./...

generate: statik
	( \
		cd internal/krmfunction; \
		statik -src=$(FUNC_WRAPPER_SRC_DIR) -f -p $(FUNC_WRAPPER_DST_DIR) -include=main.go,go.mod.src \
	)

build: generate
	go build -o pluginator main.go

install: generate
	go install .

run: generate
	go run . $(ARGS)

clean:
	rm -f pluginator
