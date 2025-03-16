FROM gcr.io/distroless/static-debian12
ARG binary
LABEL maintainer="Jan Delgado <jdelgado@gmx.net>"

COPY $binary /rabtap
ENTRYPOINT ["/rabtap"]