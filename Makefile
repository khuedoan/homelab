.POSIX:

default: apply

.PHONY: metal
metal:
	make -C metal

.PHONY: cluster
cluster:
	make -C cluster

.PHONY: apps
apps:
	make -C apps

.PHONY: tools
tools:
	make -C tools

.PHONY: docs
docs:
	make -C docs

.PHONY: apply
apply: metal cluster apps

lint:
	# TODO (feature) Add lint checks for everything
	make -C metal lint
	make -C cluster lint

hooks:
	cp ./scripts/hooks/* .git/hooks/
