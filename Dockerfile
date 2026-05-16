# Runtime image for the tene CLI.
#
# Built by GoReleaser: the `dockers:` step drops the cross-compiled binary
# into the build context root, so this Dockerfile simply copies it into a
# minimal Alpine base. No Go toolchain in the final image.

FROM alpine:3.19 AS runtime

RUN apk add --no-cache ca-certificates \
 && addgroup -S tene \
 && adduser -S -G tene tene

COPY tene /usr/local/bin/tene

USER tene
WORKDIR /workspace

ENTRYPOINT ["/usr/local/bin/tene"]
CMD ["--help"]

LABEL org.opencontainers.image.source="https://github.com/agent-kay-it/tene"
LABEL org.opencontainers.image.description="Local-first encrypted secret manager CLI for AI-safe workflows"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.vendor="agent-kay-it"
