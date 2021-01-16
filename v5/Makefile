GOFILES := $(shell find . -name "*.go")

VERSION := "v5"

# Check if any .go files need to be reformatted.
.PHONY: fmt-check
fmt-check:
	@diff=$$(gofmt -s -d $(GOFILES)); \
	if [ -n "$$diff" ]; then \
		echo "$${diff}"; \
		exit 1; \
	fi;

# Run golint across all .go files. A confidence interval of 0.3 will not error out when files in the package don't have
# a standard package header comment.
.PHONY: lint-check
lint-check:
	@for file in $(GOFILES); do \
		golint -min_confidence 0.3 $$file || exit 1; \
	done;

# Run the tests.
.PHONY: test
test:
	@if [ ! -d $(VERSION) ]; then \
		echo "Invalid version: $(VERSION)"; \
		exit 1; \
	fi; \
	cd $(VERSION); \
	go test; \
	cd ..;
