.POSIX:

filter=.

default: test

test:
	gotestsum --format testname -- -timeout 30m -run "${filter}"
