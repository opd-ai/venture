#! /usr/bin/env sh

# Number of iterations to perform
# (must be a positive integer)
ITER=30


echo "iteration $i started."
copilot -p "/delegate $(cat docs/FIX.md)" --allow-all-tools --deny-tool sudo
make fmt
echo "iteration $i in progress."
echo "Fix completed, sleeping for 1 minute..."
sleep 1m
copilot -p "/delegate $(cat docs/CHECKIN.md)" --allow-all-tools --deny-tool sudo
echo "iteration $i in progress."
echo "Fix checkin completed, sleeping for 1 minute..."
sleep 1m

for i in $(seq 1 $ITER); do
    echo "iteration $i started."
    copilot -p "/delegate $(cat docs/AUTO.md)" --allow-all-tools --deny-tool sudo
    make fmt
    echo "iteration $i in progress."
    echo "Auto fix completed, sleeping for 1 minute..."
    sleep 1m
    copilot -p "/delegate $(cat docs/CHECKIN.md)" --allow-all-tools --deny-tool sudo
    echo "iteration $i in progress."
    echo "Auto fix checkin completed, sleeping for 1 minute..."
    sleep 1m

    if [ $((i % 5)) -eq 0 ]; then
        copilot -p "/delegate $(cat docs/PERFORMANCE.md)" --allow-all-tools --deny-tool sudo
        make fmt
        echo "iteration $i in progress."
        echo "Performance completed, sleeping for 1 minute..."
        sleep 1m
        copilot -p "/delegate $(cat docs/CHECKIN.md)" --allow-all-tools --deny-tool sudo
        echo "iteration $i in progress."
        echo "Performance checkin completed, sleeping for 1 minute..."
        sleep 1m
        echo "iteration $i in complete."
    fi

    copilot -p "/delegate $(cat docs/BREAKDOWN.md)" --allow-all-tools --deny-tool sudo
    make fmt
    echo "iteration $i in progress."
    echo "Breakdown completed, sleeping for 1 minute..."
    sleep 1m
    copilot -p "/delegate $(cat docs/CHECKIN.md)" --allow-all-tools --deny-tool sudo
    echo "iteration $i in progress."
    echo "Breakdown checkin completed, sleeping for 1 minute..."
    sleep 1m
    echo "iteration $i in complete."
done