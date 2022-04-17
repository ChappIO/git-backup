FROM alpine:3

ARG TARGETPLATFORM

RUN apk add --no-cache libc6-compat

ADD ${TARGETPLATFORM}/git-backup /
RUN chmod +x /git-backup

VOLUME /backups

ENTRYPOINT ["/git-backup", "-backup.path", "/backups", "-config.file", "/backups/git-backup.yml"]
