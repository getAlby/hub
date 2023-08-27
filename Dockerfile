FROM golang:1.20-alpine as builder

# Move to working directory /build
WORKDIR /build

# Copy and download dependency using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container
COPY . .

# Build the application
RUN go build -o main

# Start a new, final image to reduce size.

FROM gcr.io/distroless/static-debian11

USER small-user:small-user

# Copy the binaries and entrypoint from the builder image.
COPY --from=builder /build/main /bin/
COPY --from=builder /build/public /public/
COPY --from=builder /build/views /views/

ENTRYPOINT [ "/bin/main" ]
