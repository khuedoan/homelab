.POSIX:

default: book

mermaid*.js:
	mdbook-mermaid install .

.PHONY: book
book: mermaid*.js
	mdbook build .
