.POSIX:
.PHONY: default infra smoke-test test lint docs fmt
.EXPORT_ALL_VARIABLES:

env ?=

default: infra

infra:
	@test -n "$(env)" || { echo 'Usage: make infra env=production (or staging)' >&2; exit 1; }
	cd "infra/$(env)" && terragrunt run --all apply

smoke-test:
	make -C tests e2e filter='^Apps$$' config='$(TEST_CONFIG)'

test:
	make -C toolbox test
	make -C tests test

docs:
	mkdocs serve

fmt:
	treefmt
