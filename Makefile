.PHONY: all
all: build
	gpg --batch --yes --output ./dist/*_SHA256SUMS.sig --detach-sign ./dist/*_SHA256SUMS

.PHONY: build
build:
	goreleaser --skip publish --clean