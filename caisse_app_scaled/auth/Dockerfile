FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY out/auth/ .

EXPOSE 8092

# Set environment variables
ENV GATEWAY=172.17.0.1

# Run the application
CMD ["./app"]