#! /usr/bin/env sh

# Number of iterations to perform
# (must be a positive integer)
ITER=15

alias copilot="yes n | copilot --model claude-sonnet-4.5"

#./cmd.sh impl
rm -fv docs/PLAN.md docs/INTEGRATION_AUDIT.md
./cmd.sh integrate
./cmd.sh fix
./cmd.sh perf
./cmd.sh perf
./cmd.sh perf
./cmd.sh perf
./cmd.sh perf
#./cmd.sh wasm
#./cmd.sh mobile
for i in $(seq 1 $ITER); do
    ./cmd.sh dev
done

git clean -fdx .
./cmd.sh fix