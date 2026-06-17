.PHONY: all
all: build

.PHONY: docs
docs:
	go generate ./...

.PHONY: release
release:
	goreleaser release --clean
