FROM alpine:latest

# Set build-time variable, imported from pipeline environment during `Build` step
ARG IMAGE_NAME
ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH

# Copy the ARG value to an ENV variable that will persist at runtime
ENV IMAGE_NAME=${IMAGE_NAME}

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create a non-root user with a fixed UID and group ID
RUN addgroup -g 1000 ${IMAGE_NAME} && \
    adduser -D -u 1000 -G ${IMAGE_NAME} ${IMAGE_NAME}

# Copy the entire dist directory
COPY dist/ /tmp/dist/

# Copy the correct binary based on target platform
# GoReleaser creates separate directories for each platform without no_unique_dist_dir
RUN if [ "$TARGETOS" = "linux" ] && [ "$TARGETARCH" = "amd64" ]; then \
        BINARY_DIR="${IMAGE_NAME}_linux_amd64*"; \
    elif [ "$TARGETOS" = "linux" ] && [ "$TARGETARCH" = "arm64" ]; then \
        BINARY_DIR="${IMAGE_NAME}_linux_arm64*"; \
    else \
        echo "Unsupported platform: $TARGETOS/$TARGETARCH" && exit 1; \
    fi && \
    DIST_DIR=$(find /tmp/dist -name "$BINARY_DIR" -type d | head -1) && \
    cp "$DIST_DIR/${IMAGE_NAME}" /usr/local/bin/${IMAGE_NAME} && \
    chmod +x /usr/local/bin/${IMAGE_NAME} && \
    chown ${IMAGE_NAME}:${IMAGE_NAME} /usr/local/bin/${IMAGE_NAME} && \
    rm -rf /tmp/dist

# Switch to the non-root user and set working directory
USER ${IMAGE_NAME}
WORKDIR /home/${IMAGE_NAME}

# Use JSON array format for ENTRYPOINT with shell to allow variable substitution
ENTRYPOINT ["/bin/sh", "-c", "/usr/local/bin/${IMAGE_NAME}"]