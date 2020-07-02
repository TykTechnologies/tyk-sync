FROM ruby:2

RUN gem install package_cloud
COPY pc-wrapper.sh /pc-wrapper.sh

ENTRYPOINT [ "/pc-wrapper.sh" ]
CMD []
