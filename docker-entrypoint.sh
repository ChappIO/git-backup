#!/bin/sh
set -eu
if [[ -z "${PUID-}" && -z "${PGID-}" ]]
then
  # We are running through normal docker user changes, so nothing special to do
  /git-backup "$@"
else
  # We are running with an environment variable user change
  PUID=${PUID:-$(id -u)}
  PGID=${PGID:-$(id -g)}

  # Make sure the user exists
  useradd -o -u "$PUID" -U -d /backups -s /bin/false git-backup
  groupmod -o -g "$PGID" git-backup

  # Own the backups folder
  chown git-backup:git-backup /backups

  # Let's go!
  echo /git-backup "$@" | su -s /bin/sh git-backup
fi

