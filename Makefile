.SILENT:

killall:
	for pid in $$(ps | grep -E \
		'skipper$$' \
		| sed 's/ .*//'); do \
		kill $$pid; \
	done

step0: killall
	skipper &
