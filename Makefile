.POSIX:

default: metal bootstrap

.PHONY: metal
metal:
	make -C metal

.PHONY: bootstrap
bootstrap:
	make -C bootstrap

.PHONY: tools
tools:
	make -C tools

.PHONY: docs
docs:
	make -C docs

lint:
	# TODO (feature) Add lint checks for everything
	make -C metal lint
