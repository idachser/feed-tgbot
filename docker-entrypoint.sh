#!/bin/sh
set -eu

mkdir -p /app
chown -R bot:bot /app

exec su-exec bot "$@"
