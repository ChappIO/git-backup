#!/bin/sh
set -eu

PUID=${PUID:-1000}
PGID=${PGID:-1000}

groupmod -o -g "$PGID" git-backup
usermod -o -u "$PUID"  git-backup

/git-backup "$@"
