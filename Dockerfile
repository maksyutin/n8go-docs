# ─────────────────────────────────────────────────────────────────────────────
# Multi-stage Dockerfile for n8go-docs
#
# The binary is built by the CI pipeline and injected via COPY.
# This keeps the image tiny (~10 MB) and the build layer cache effective.
# ─────────────────────────────────────────────────────────────────────────────

ARG BINARY_NAME=n8go-docs

# ── Runtime image ─────────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot AS runtime

ARG BINARY_NAME
ARG BUILD_DATE
ARG VCS_REF
ARG VERSION

LABEL org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.revision="${VCS_REF}" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.title="n8go-docs" \
      org.opencontainers.image.description="Static documentation generator"

# Copy the pre-built binary from the CI artifact
COPY dist/${BINARY_NAME} /usr/local/bin/n8go-docs

# Docs and themes directories are expected to be mounted as volumes
VOLUME ["/docs", "/site"]

# Health endpoint is served by the app on /health when running in serve mode
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/n8go-docs"]
CMD ["serve", "--port", "8080"]
