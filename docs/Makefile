.POSIX:

todo:
	printf "# TODO\n\n" > todo.md
	git grep --line-number TODO .. ':!.' \
		| awk --field-separator ':| TODO ' '{ printf "- [%s](%s#L%s)\n", $$4, $$1, $$2 }' \
		| sort \
		>> todo.md
