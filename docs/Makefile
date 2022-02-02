.POSIX:

default: book

.PHONY: diagrams
diagrams:
	cd diagrams \
		&& python3 -m venv .venv \
		&& source .venv/bin/activate \
		&& pip3 install -r requirements.txt \
		&& python3 *.py
	mkdir -p src/images
	cp diagrams/*.png src/images/

.PHONY: book
book: diagrams
	mdbook build .
