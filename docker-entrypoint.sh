#!/bin/sh
set -eu

PUID=${PUID:-$(id -u)}
PGID=${PGID:-$(id -g)}

groupmod -o -g "$PGID" git-backup
usermod -o -u "$PUID"  git-backup
chown git-backup:git-backup /backup

su git-backup /git-backup "$@"
