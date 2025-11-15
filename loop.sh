#! /usr/bin/env sh

copilot -p "/delegate $(cat docs/FIX.md)" --allow-all-tools --deny-tool sudo
copilot -p "/delegate $(cat docs/EXECUTE.md)" --allow-all-tools --deny-tool sudo
copilot -p "/delegate $(cat docs/CHECKIN.md)" --allow-all-tools --deny-tool sudo
sleep 5m
