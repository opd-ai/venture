#! /usr/bin/env sh

# Number of iterations to perform
# (must be a positive integer)
ITER=6

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
    copilot -p "/delegate $(cat docs/EXECUTE.md)" --allow-all-tools --deny-tool sudo
    make fmt
    echo "iteration $i in progress."
    echo "Execute completed, sleeping for 1 minute..."
    sleep 1m
    copilot -p "/delegate $(cat docs/CHECKIN.md)" --allow-all-tools --deny-tool sudo
    echo "iteration $i in progress."
    echo "Execute checkin completed, sleeping for 1 minute..."
    sleep 1m
    echo "iteration $i in complete."

    copilot -p "/delegate $(cat docs/INTEGRATION.md)" --allow-all-tools --deny-tool sudo
    make fmt
    echo "iteration $i in progress."
    echo "Integrate completed, sleeping for 1 minute..."
    sleep 1m
    copilot -p "/delegate $(cat docs/CHECKIN.md)" --allow-all-tools --deny-tool sudo
    echo "iteration $i in progress."
    echo "Integrate checkin completed, sleeping for 1 minute..."
    sleep 1m
    echo "iteration $i in complete."

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
done

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