.POSIX:

default: book

.PHONY: diagrams
diagrams:
	cd diagrams \
		&& python3 -m venv .venv \
		&& pip3 install -r requirements.txt \
		&& python3 *

.PHONY: book
book:
	mdbook build .
