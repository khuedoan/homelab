.POSIX:

all: metal infra apps

.PHONY: metal
metal:
	make -C metal

.PHONY: infra
infra:
	make -C infra

.PHONY: apps
apps:
	make -C apps
