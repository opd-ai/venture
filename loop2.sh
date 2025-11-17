#! /usr/bin/env sh

# Number of iterations to perform
# (must be a positive integer)
ITER=1

fix() { 
    go test -race ./...
    # if the tests failed, run this script only once to attempt to fix the issues
    if [ $? -ne 0 ]; then
        echo "iteration started - Fix phase."
        copilot -p "/delegate $(cat docs/FIX.md)" --allow-all-tools --deny-tool sudo
        make fmt
        echo "Fix phase completed, sleeping for 1 minute..."
        sleep 1m
        checkin
    fi
}

checkin() {
    git diff --quiet
    # If there are changes, run this script to check them in
    if [ $? -ne 0 ]; then
        echo "iteration started - Checkin phase."
        copilot -p "/delegate $(cat docs/CHECKIN.md)" --allow-all-tools --deny-tool sudo
        echo "Checkin phase completed, sleeping for 1 minute..."
        sleep 1m
    fi
}

dev() {
    echo "iteration started - Maintenance phase 1: Functional Breakdown."
    copilot -p "/delegate $(cat docs/BREAKDOWN.md)" --allow-all-tools --deny-tool sudo
    echo "Implementation in progress."
    make fmt
    echo "Implementation completed, sleeping for 1 minute..."
    sleep 1m
    checkin

    echo "iteration started - Maintenance phase 2: General components."
    copilot -p "/delegate $(cat docs/AUTO.md)" --allow-all-tools --deny-tool sudo
    echo "Integration in progress."
    make fmt
    echo "Integration completed, sleeping for 1 minute..."
    sleep 1m
    checkin

    echo "iteration started - Maintenance Phase 3: Resolution check."
    copilot -p "/delegate $(cat docs/PLAY.md)" --allow-all-tools --deny-tool sudo
    make fmt
    echo "Resolution check completed, sleeping for 1 minute..."
    sleep 1m
    checkin
}

fix

for i in $(seq 1 $ITER); do
    dev
    fix
done

fix