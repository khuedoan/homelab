.POSIX:

default: tools

.PHONY: tools
tools:
	make -C tools

.PHONY: deploy
deploy:
	make -C metal reset
	make -C infra
	make -C apps

.PHONY: apply
apply:
	make -C metal
	make -C infra
	make -C apps

.PHONY: docs
docs:
	make -C docs
