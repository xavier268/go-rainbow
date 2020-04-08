#!/bin/bash

FILE=./prof/*/cpu*

echo "Profiling the demo program, and saving results"
go run ./cmd/

FILE=$(ls -1 $FILE | sort -r | head -n 1)
echo "Saved profiling data in $FILE"

# open profile call tree in web brower
go tool pprof -web $FILE

