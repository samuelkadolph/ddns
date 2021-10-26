# Build Stage
FROM golang:bullseye AS build

WORKDIR /go/src/github.com/samuelkadolph/ddns

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 make clean test build

# Run Stage
FROM alpine:latest

MAINTAINER samuel@kadolph.com

ARG BUILD_DATE
ARG VCS_REF

LABEL org.label-schema.build-date=$BUILD_DATE
LABEL org.label-schema.name="ifconfig"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.vcs-ref=$VCS_REF
LABEL org.label-schema.vcs-url="https://github.com/samuelkadolph/ddns"

WORKDIR /root

COPY --from=build /go/src/github.com/samuelkadolph/ddns/build/ddns .

EXPOSE 4444

CMD ["./ddns"]
