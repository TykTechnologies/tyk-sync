FROM debian:buster-slim

ARG TYKVERSION

LABEL Description="Tyk Sync docker image" Vendor="Tyk" Version=$TYKVERSION

RUN apt-get update \
 && apt-get upgrade -y \
 && apt-get install -y --no-install-recommends \
            curl ca-certificates apt-transport-https debian-archive-keyring gnupg \
 && curl -L https://packagecloud.io/tyk/tyk-sync/gpgkey | apt-key add - \
 && apt-get purge -y gnupg \
 && apt-get autoremove -y \
 && rm -rf /root/.cache

RUN echo "deb https://packagecloud.io/tyk/tyk-sync/debian/ buster main" | tee /etc/apt/sources.list.d/tyk_tyk-sync.list \
 && apt-get update \
 && apt-get install -y tyk-sync=$TYKVERSION \
 && rm -rf /var/lib/apt/lists/*

 WORKDIR /opt/tyk-sync

 ENTRYPOINT ["./tyk-sync"]
