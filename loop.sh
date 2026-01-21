#! /usr/bin/env sh
git commit -am "audit"
git push origin main
git pull origin main --no-rebase;
devloop.sh
sleep 1m
breakloop.sh
