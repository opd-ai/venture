#! /usr/bin/env sh

while true; do
    copilot -p "/delegate $(cat docs/FIX.md)" --allow-all-tools --deny-tool sudo
    echo "Fix completed, sleeping for 1 minute..."
    sleep 1m
    copilot -p "/delegate $(cat docs/CHECKIN.md)" --allow-all-tools --deny-tool sudo
    echo "Fix checkin completed, sleeping for 1 minute..."
    sleep 1m
    copilot -p "/delegate $(cat docs/EXECUTE.md)" --allow-all-tools --deny-tool sudo
    echo "Execute completed, sleeping for 1 minute..."
    sleep 1m
    copilot -p "/delegate $(cat docs/CHECKIN.md)" --allow-all-tools --deny-tool sudo
    echo "Execute checkin completed, sleeping for 1 minute..."
    sleep 1m
done