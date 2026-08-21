FROM gcr.io/distroless/static-debian13@sha256:f2ea2709ac8db56323cbd7d014277f32cb572d9ea124b0076f7aafe5980678fe
ARG binary
LABEL maintainer="Jan Delgado <jdelgado@gmx.net>"

COPY $binary /rabtap
ENTRYPOINT ["/rabtap"]
