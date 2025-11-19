#! /usr/bin/env sh

# Number of iterations to perform
# (must be a positive integer)
ITER=200

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

auto() {
    echo "iteration started - Maintenance Phase 3: Auto check."
    copilot -p "/delegate $(cat docs/AUTO.md)" --allow-all-tools --deny-tool sudo
    make fmt
    echo "Auto check completed, sleeping for 1 minute..."
    sleep 1m
    checkin
}

play() {
    echo "iteration started - Maintenance Phase 4: Resolution check."
    copilot -p "/delegate $(cat docs/PLAY.md)" --allow-all-tools --deny-tool sudo
    make fmt
    echo "Resolution check completed, sleeping for 1 minute..."
    sleep 1m
    checkin
}

log() {
    echo "iteration started - Maintenance Phase 5: Logging."
    copilot -p "/delegate $(cat docs/LOG.md)" --allow-all-tools --deny-tool sudo
    make fmt
    echo "Logging completed, sleeping for 1 minute..."
    sleep 1m
    checkin
}

dev() {
    echo "iteration started - Maintenance phase 1: Functional Breakdown."
    copilot -p "/delegate $(cat docs/BREAKDOWN.md)" --allow-all-tools --deny-tool sudo
    echo "Implementation in progress."
    make fmt
    echo "Implementation completed, sleeping for 1 minute..."
    sleep 1m
    checkin

    #check if any directories do not contain an AUDIT.md file, if so, set NEED_AUDIT to true
    NEED_AUDIT=false
    for dir in */; do
        if [ ! -f "$dir/AUDIT.md" ]; then
            NEED_AUDIT=true
            break
        fi
    done

    if [ "$NEED_AUDIT" = true ]; then
        echo "iteration started - Maintenance phase 2: General components."
        copilot -p "/delegate $(cat docs/REVIEW.md)" --allow-all-tools --deny-tool sudo
        echo "Review in progress."
        make fmt
        echo "Review completed, sleeping for 1 minute..."
        sleep 1m
        checkin
    else
        echo "All directories have AUDIT.md files. Skipping Audit phase."
    fi
}

fix

for i in $(seq 1 $ITER); do
    log
    dev
    fix
done

fix