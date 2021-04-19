.POSIX:

todo:
	printf "# TODO\n\n" > TODO.md
	git grep -h TODO .. ':!.' | sed 's/.* TODO /- /g' >> TODO.md
