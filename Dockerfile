FROM ruby:4.0-alpine

ARG BUILD_DATE
ARG VCS_REF
ARG VERSION

RUN apk update && apk --no-cache --update add build-base

ENV BUNDLE_WITHOUT=development
ENV RACK_ENV=production

WORKDIR /app

COPY Gemfile .
COPY Gemfile.lock .

RUN gem install bundler && bundle install

COPY . .

LABEL org.opencontainers.image.authors="samuel@kadolph.com"
LABEL org.opencontainers.image.created="$BUILD_DATE"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.revision="$VCS_REF"
LABEL org.opencontainers.image.source="https://github.com/samuelkadolph/ddns"
LABEL org.opencontainers.image.title="ddns"
LABEL org.opencontainers.image.url="https://github.com/samuelkadolph/ddns"
LABEL org.opencontainers.image.version="$VERSION"

EXPOSE 7860

CMD ["bundle", "exec", "puma", "--bind", "tcp://0.0.0.0:7860"]

