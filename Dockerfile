FROM alpine:3

ARG TARGETPLATFORM

VOLUME /backups

RUN apk add --no-cache libc6-compat

ADD ${TARGETPLATFORM}/git-backup /
RUN chmod +x /git-backup

## Add the user for command execution
RUN apk add --no-cache shadow
RUN groupmod -g 0 users && \
 useradd -o -u 0 -U -d /backups -s /bin/false git-backup && \
 usermod -G users git-backup

ADD ./docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

ENTRYPOINT ["/docker-entrypoint.sh", "-backup.path", "/backups", "-config.file", "/backups/git-backup.yml"]
