FROM golang:1.20-alpine as builder

# Move to working directory /build
WORKDIR /build

# Copy and download dependency using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container
COPY . .

# TODO: build react app?

# Build the application
RUN go build -o main

# Start a new, final image to reduce size.
FROM alpine as final

# Copy the binaries and entrypoint from the builder image.
COPY --from=builder /build/main /bin/
# NOTE: should not be needed - assets should be embedded in the go app
#COPY --from=builder /build/public /public/
#COPY --from=builder /build/views /views/

ENTRYPOINT [ "/bin/main" ]
