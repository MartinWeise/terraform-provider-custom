.PHONY: all
all: build

.PHONY: build
build:
	goreleaser release --clean
