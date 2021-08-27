#!/bin/bash

while true; do
	echo "Recovering..."
	make run >/dev/null 2>&1
	sleep 1
done
