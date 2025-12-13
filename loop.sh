#! /usr/bin/env sh

# Number of iterations to perform
# (must be a positive integer)
ITER=15

alias copilot="yes n | copilot --model claude-sonnet-4.5"

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

wasm() {
    echo "iteration started - Resolution phase."
    copilot -p "/delegate $(cat docs/PLAY-WASM.md)" --allow-all-tools --deny-tool sudo
    make fmt
    echo "Resolution phase completed, sleeping for 1 minute..."
    sleep 1m
    checkin
}

mobile() {
    echo "iteration started - Resolution phase."
    copilot -p "/delegate $(cat docs/PLAY-MOBILE.md)" --allow-all-tools --deny-tool sudo
    make fmt
    echo "Resolution phase completed, sleeping for 1 minute..."
    sleep 1m
    checkin
}

impl() {
    echo "iteration started - Dev phase 1: Implement changes."
    copilot -p "/delegate $(cat docs/EXECUTE.md)" --allow-all-tools --deny-tool sudo
    echo "Implementation in progress."
    make fmt
    echo "Implementation completed, sleeping for 1 minute..."
    sleep 1m
    checkin
}
review() {
    echo "iteration started - Review phase."
    copilot -p "/delegate $(cat docs/REVIEW.md)" --allow-all-tools --deny-tool sudo
    make fmt
    echo "Review completed, sleeping for 1 minute..."
    sleep 1m
    checkin   
}
integrate() {
    echo "iteration started - Dev phase 2: Integrate components."
    copilot -p "/delegate $(cat docs/INTEGRATION.md)" --allow-all-tools --deny-tool sudo
    echo "Integration in progress."
    make fmt
    echo "Integration completed, sleeping for 1 minute..."
    sleep 1m
    checkin
}
play() {
    echo "iteration started - Resolution phase."
    copilot -p "/delegate $(cat docs/PLAY.md)" --allow-all-tools --deny-tool sudo
    make fmt
    echo "Resolution phase completed, sleeping for 1 minute..."
    sleep 1m
    checkin   
}

dev() {
    impl
    review
    integrate
    #play
}

fix
#wasm
#mobile
for i in $(seq 1 $ITER); do
    dev
    fix
done

git clean -fdx .
fix