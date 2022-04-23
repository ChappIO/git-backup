#!/bin/sh
set -eu
if [[ -v $PUID || -v $PGID ]]
then
  # We are running with an environment variable user change
  PUID=${PUID:-$(id -u)}
  PGID=${PGID:-$(id -g)}

  # Make sure the user exists
  useradd -o -u "$PUID" -U -d /backups -s /bin/false git-backup
  groupmod -o -g "$PGID" git-backup

  # Own the backups folder
  chown git-backup:git-backup /backup

  # Let's go!
  su -s /bin/sh git-backup -c "/git-backup $@" whoami
else
  # We are running through normal docker user changes, so nothing special to do
  git-backup /git-backup "$@"
fi

