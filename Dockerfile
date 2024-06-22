FROM node:20-alpine as frontend
WORKDIR /build
COPY frontend ./frontend
RUN cd frontend && yarn install && yarn build:http

FROM golang:1.22.2 as builder

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
   -ldflags="-X 'github.com/getAlby/nostr-wallet-connect/version.Tag=$TAG'" \
   -o main cmd/http/main.go

RUN cp `find /go/pkg/mod/github.com/breez/ |grep linux-amd64 |grep libbreez_sdk_bindings.so` ./
RUN cp `find /go/pkg/mod/github.com/get\!alby/ | grep x86_64-unknown-linux-gnu | grep libglalby_bindings.so` ./
RUN cp `find /go/pkg/mod/github.com/get\!alby/ | grep x86_64-unknown-linux-gnu | grep libldk_node.so` ./

# Start a new, final image to reduce size.
FROM debian as final

ENV LD_LIBRARY_PATH=/usr/lib/nwc
#
# # Copy the binaries and entrypoint from the builder image.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/libbreez_sdk_bindings.so /usr/lib/nwc/
COPY --from=builder /build/libglalby_bindings.so /usr/lib/nwc/
COPY --from=builder /build/libldk_node.so /usr/lib/nwc/
COPY --from=builder /build/main /bin/

ENTRYPOINT [ "/bin/main" ]
