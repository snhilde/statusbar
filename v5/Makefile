GOFILES := $(shell find . -name "*.go")

# Check if any .go files need to be reformatted.
.PHONY: fmt-check
fmt-check:
	@diff=$$(gofmt -s -d $(GOFILES)); \
	if [ -n "$$diff" ]; then \
		echo "$${diff}"; \
		exit 1; \
	fi;

# Run a large number of linters of various types and purposes across all go source files.
.PHONY: lint-check-source
lint-check-source:
	@if [ ! -f .golangci.yml ]; then \
		echo "Missing .golangci.yml"; \
		exit 1; \
	fi; \
	golangci-lint run --skip-files ".*_test.*";

# Run a large number of linters of various types and purposes across all go files (including test files).
.PHONY: lint-check-all
lint-check-all:
	@if [ ! -f .golangci.yml ]; then \
		echo "Missing .golangci.yml"; \
		exit 1; \
	fi; \
	golangci-lint run;

# Run the tests.
.PHONY: test
test:
	go test || exit $?;
