.EXPORT_ALL_VARIABLES:

GO111MODULE ?= on
LOCALS      := $(shell find . -type f -name '*.go')

TOOLS       := $(wildcard cmd/*)
.PHONY: fmt deps test build $(TOOLS)

all: fmt deps test build

deps:
	go get ./...

fmt:
	gofmt -w $(LOCALS)
	go vet ./...

test:
	go test -count=1 ./...

$(TOOLS):
	go build -o $(subst cmd,bin,$(@)) $(@)/*.go

build: $(TOOLS)