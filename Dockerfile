FROM node:20-alpine AS frontend

# Set the base path for the frontend build
# This can be overridden at build time with --build-arg BASE_PATH=<url> e.g. --build-arg BASE_PATH=/hub
# Allows to build a frontend that can be served from a subpath, e.g. /hub
ARG BASE_PATH
WORKDIR /build
COPY frontend ./frontend
RUN echo "Building frontend with base path $BASE_PATH"
RUN cd frontend && yarn install --network-timeout 3000000 && yarn build:http

FROM golang:1.24 AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TAG

RUN echo "I am running on $BUILDPLATFORM, building for $TARGETPLATFORM, release tag $TAG"

RUN apt-get update && \
   apt-get install -y gcc

ENV CGO_ENABLED=1
ENV GOOS=linux
#ENV GOARCH=$GOARCH

#RUN echo "AAA $GOARCH"

# Move to working directory /build
WORKDIR /build

# Copy and download dependency using go mod
COPY go.mod .
COPY go.sum .
RUN GOARCH=$(echo "$TARGETPLATFORM" | cut -d'/' -f2) go mod download

# Copy the code into the container
COPY . .

# Copy frontend dist files into the container
COPY --from=frontend /build/frontend/dist ./frontend/dist

RUN GOARCH=$(echo "$TARGETPLATFORM" | cut -d'/' -f2) go build \
   -ldflags="-X 'github.com/getAlby/hub/version.Tag=$TAG'" \
   -o main cmd/http/main.go

RUN GOARCH=$(echo "$TARGETPLATFORM" | cut -d'/' -f2) go build \
   -o db_migrate cmd/db_migrate/main.go

COPY ./build/docker/copy_dylibs.sh .
RUN chmod +x copy_dylibs.sh
RUN ./copy_dylibs.sh $(echo "$TARGETPLATFORM" | cut -d'/' -f2)

# Start a new, final image to reduce size.
FROM debian:12-slim AS final

ENV LD_LIBRARY_PATH=/usr/lib/nwc
#
# # Copy the binaries and entrypoint from the builder image.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/libldk_node.so /usr/lib/nwc/
COPY --from=builder /build/main /bin/
COPY --from=builder /build/db_migrate /bin/

ENTRYPOINT [ "/bin/main" ]
