.POSIX:

default: metal bootstrap

all: default external

.PHONY: metal
metal:
	make -C metal

.PHONY: bootstrap
bootstrap:
	make -C bootstrap

.PHONY: external
external:
	make -C external

.PHONY: tools
tools:
	make -C tools

.PHONY: docs
docs:
	make -C docs

dev:
	make -C metal cluster env=dev
	make -C bootstrap
