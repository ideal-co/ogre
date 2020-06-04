FROM alpine:latest

LABEL ogre.health.check.health.script='/usr/local/bin/health.sh'

LABEL ogre.format.backend.prometheus="true"

VOLUME ./health.sh:/usr/local/bin/health.sh

ENTRYPOINT ["/bin/sh"]
