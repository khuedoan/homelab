.POSIX:

default: metal bootstrap wait

all: metal bootstrap external wait

configure:
	./scripts/configure
	git status

.PHONY: metal
metal:
	make -C metal

.PHONY: bootstrap
bootstrap:
	make -C bootstrap

.PHONY: external
external:
	make -C external

wait:
	./scripts/wait-main-apps

.PHONY: tools
tools:
	make -C tools

.PHONY: docs
docs:
	make -C docs

dev:
	make -C metal cluster env=dev
	make -C bootstrap
