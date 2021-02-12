.POSIX:

all: metal infra platform apps

.PHONY: metal
metal:
	make -C metal

.PHONY: infra
infra:
	make -C infra

.PHONY: platform
platform:
	make -C platform

.PHONY: apps
apps:
	make -C apps
