FROM debian:buster-slim

ARG TYKVERSION
ARG REPOSITORY
LABEL Description="Tyk Sync docker image" Vendor="Tyk" Version=$TYKVERSION

RUN apt-get update \
 && apt-get upgrade -y \
 && apt-get install -y --no-install-recommends \
            curl ca-certificates apt-transport-https debian-archive-keyring gnupg \
 && curl -L https://packagecloud.io/tyk/$REPOSITORY/gpgkey | apt-key add - \
 && apt-get purge -y gnupg \
 && apt-get autoremove -y \
 && rm -rf /root/.cache

RUN echo "deb https://packagecloud.io/tyk/$REPOSITORY/debian/ jessie main" | tee /etc/apt/sources.list.d/tyk_tyk-sync.list \
 && apt-get update \
 && apt-get install -y tyk-sync=$TYKVERSION \
 && rm -rf /var/lib/apt/lists/*
