.POSIX:

default: book

.PHONY: diagrams
diagrams:
	cd diagrams \
		&& python *

.PHONY: book
book:
	mdbook build .
