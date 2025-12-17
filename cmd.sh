#! /usr/bin/env sh

# Number of iterations to perform
# (must be a positive integer)
ITER=15

alias copilot="yes n | copilot --model claude-opus-4.5"

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
perf() {
    echo "iteration started - Dev phase 2: Optimize performance."
    copilot -p "/delegate $(cat docs/PERFORMANCE.md)" --allow-all-tools --deny-tool sudo
    echo "Optimization in progress."
    make fmt
    echo "Optimization completed, sleeping for 1 minute..."
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
auto() {
    echo "iteration started - Maintenance Phase 3: Auto check."
    copilot -p "/delegate $(cat docs/AUTO.md)" --allow-all-tools --deny-tool sudo
    make fmt
    echo "Auto check completed, sleeping for 1 minute..."
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

base_audit() {
    echo "iteration started - Maintenance phase 1: Base components."
    copilot -p "/delegate $(cat docs/BASE_AUDIT.md)" --allow-all-tools --deny-tool sudo
    make fmt
    echo "Base component audit completed, sleeping for 1 minute..."
    sleep 1m
    checkin
}

audit3() {
    echo "iteration started - Maintenance phase 3: Tertiary components."
    copilot -p "/delegate $(cat docs/audit3.md)" --allow-all-tools --deny-tool sudo
    make fmt
    echo "Tertiary component audit completed, sleeping for 1 minute..."
    sleep 1m
    checkin
}

audit() {
    #check if any directories do not contain an AUDIT.md file, if so, set NEED_AUDIT to true
    NEED_AUDIT=false
    DIR="pkg/"
    for dir in pkg/*/; do
        if [ ! -f "$dir/AUDIT.md" ]; then
            NEED_AUDIT=true
            DIR="$dir"
            break
        fi
    done

    if [ "$NEED_AUDIT" = true ]; then
        echo "iteration started - Maintenance phase 2: General components."
        copilot -p "/delegate The \`$DIR\` is known to be un-audited. $(cat docs/REVIEW.md)" --allow-all-tools --deny-tool sudo
        echo "Review in progress."
        make fmt
        echo "Review completed, sleeping for 1 minute..."
        sleep 1m
        checkin
    else
        echo "All directories have AUDIT.md files. Skipping Audit phase."
    fi
}

dev() {
    fix
    impl
    review
    perf
}

auditloop() {
    rm AUDIT.md -f
    base_audit
    audit3
}

$@