FROM gcr.io/distroless/static-debian13@sha256:3592aa8171c77482f62bbc4164e6a2d141c6122554ace66e5cc910cadb961ff0
ARG binary
LABEL maintainer="Jan Delgado <jdelgado@gmx.net>"

COPY $binary /rabtap
ENTRYPOINT ["/rabtap"]
