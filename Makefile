.PHONY: all
all: build

.PHONY: build
build:
	goreleaser release --clean

.PHONY: sign
sign:
	gpg --armor --export BD63DAC7BFB8CB531F81493D13A57764CE4AB8F0