FROM debian:buster-slim

RUN apt-get update \
 && apt-get dist-upgrade -y \
            curl ca-certificates apt-transport-https debian-archive-keyring gnupg \
 && apt-get purge -y gnupg \
 && apt-get autoremove -y \
 && rm -rf /root/.cache

COPY tyk-sync /opt/tyk-sync/tyk-sync
WORKDIR /opt/tyk-sync

ENTRYPOINT ["./tyk-sync"]

CMD ["--help"]
