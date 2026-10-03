.POSIX:
.PHONY: test e2e benchmarks

filter ?= .
override filter := $(value filter)
config ?=

test:
	go vet ./...
	go test -race -short -count=1 ./...

e2e:
	E2E=1 TEST_CONFIG='$(config)' go test -race -count=1 -timeout=30m -run '^TestCluster$$/$(filter)' -v ./e2e

benchmarks:
	E2E=1 BENCHMARKS=1 TEST_CONFIG='$(config)' go test -race -count=1 -timeout=60m -run '^TestBenchmarks$$/$(filter)' -v ./benchmarks
