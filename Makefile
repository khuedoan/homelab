.POSIX:
.PHONY: *

default: metal bootstrap wait

all: metal bootstrap external wait

configure:
	./scripts/configure
	git status

metal:
	make -C metal

bootstrap:
	make -C bootstrap

external:
	make -C external

wait:
	./scripts/wait-main-apps

tools:
	make -C tools

docs:
	make -C docs

