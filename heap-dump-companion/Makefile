GOFMT ?= gofmt -s
GOFMT_FILES?=$$(find . -name '*.go'|grep -v .cache)

build: test
	goreleaser build --snapshot --clean

package: build
	goreleaser release --skip-publish --snapshot --clean

test: fmt-check
	go generate ./...; \
    go test ./... -coverprofile=coverage.out; \
    go test ./... -json > report.json; 

fmt-check:
	@diff=$$($(GOFMT) -d $(GOFMT_FILES)); \
	if [ -n "$$diff" ]; then \
		echo "Please run 'make fmt' and commit the result:"; \
		echo "$${diff}"; \
		exit 1; \
	fi;

fmt:
	go fmt ./...

clean:
	rm -rf dist; \
	rm coverage.out; \
	rm report.json;

.PHONY: build

all:
	$(MAKE) build
	$(MAKE) package