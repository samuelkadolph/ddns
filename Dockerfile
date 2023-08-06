# Build Stage
FROM golang:bookworm AS build

WORKDIR /go/src/github.com/samuelkadolph/ddns

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 make clean test build/ddns

# Run Stage
FROM alpine:latest

MAINTAINER samuel@kadolph.com

ARG BUILD_DATE
ARG VCS_REF

WORKDIR /app

COPY --from=build /go/src/github.com/samuelkadolph/ddns/build/ddns .

LABEL org.label-schema.build-date=$BUILD_DATE
LABEL org.label-schema.name="ddns"
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.vcs-ref=$VCS_REF
LABEL org.label-schema.vcs-url="https://github.com/samuelkadolph/ddns"

EXPOSE 4444

CMD ["./ddns"]
