#! /usr/bin/env sh

for true; do
	auditor.sh
	devloop.sh
	if [ ! -f AUDIT.md ]; then
		exit 0
	fi
done
