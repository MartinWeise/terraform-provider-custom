.PHONY: all
all: build
	gpg --batch --yes --output ./dist/terraform-provider-custom_*_SHA256SUMS.sig --detach-sign ./dist/terraform-provider-custom_*_SHA256SUMS

build:
	goreleaser --skip publish --clean