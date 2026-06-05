.PHONY: all
all:
	goreleaser --skip publish
	gpg --batch --yes --output ./dist/terraform-provider-custom_*_SHA256SUMS.sig --detach-sign ./dist/terraform-provider-custom_*_SHA256SUMS