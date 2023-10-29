FROM golang:1.21-alpine as builder
ARG version

WORKDIR /go/src/app
ADD . /go/src/app

RUN    cd cmd/rabtap \
    && CGO_ENABLED=0 \
       go build -ldflags "-s -w -X main.version=$version" -o /go/bin/rabtap


FROM gcr.io/distroless/base-debian10
LABEL maintainer="Jan Delgado <jdelgado@gmx.net>"

COPY --from=builder /go/bin/rabtap /
ENTRYPOINT ["/rabtap"]
