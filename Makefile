.PHONY: fmt deps test

PKGS=`go list ./... 2> /dev/null | grep -v '/vendor'`
LOCALS=`find . -type f -name '*.go' -not -path "./vendor*/*"`

all: fmt deps test

deps:
	@go list golang.org/x/tools/cmd/goimports || go get golang.org/x/tools/cmd/goimports
	@which dep || go get -u github.com/golang/dep/cmd/dep
	#dep ensure -v

fmt:
	goimports -w $(LOCALS)
	go vet $(PKGS)

test:
	go test -test.v $(PKGS)
