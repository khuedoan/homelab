.POSIX:

default: book

.PHONY: diagrams
diagrams:
	cd diagrams \
		&& python3 -m venv .venv \
		&& pip3 install -r requirements.txt \
		&& python3 *.py

.PHONY: book
book: diagrams
	mdbook build .

.PHONY: deploy
deploy:
	@echo "Deploying to GitHub Pages"
	git worktree add /tmp/book gh-pages
	cp -rp book/* /tmp/book/
	cd /tmp/book \
		&& git commit --all --message "Updates" || echo "No changes to commit" \
		&& git push
	git worktree remove /tmp/book
