FROM node:19-alpine as frontend
WORKDIR /build
COPY frontend ./frontend
RUN cd frontend && yarn install && yarn build:http

FROM golang:1.21 as builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM

RUN echo "I am running on $BUILDPLATFORM, building for $TARGETPLATFORM"

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

RUN GOARCH=$(echo "$TARGETPLATFORM" | cut -d'/' -f2) go build -o main .

RUN cp `find /go/pkg/mod/github.com/breez/ |grep linux-amd64 |grep libbreez_sdk_bindings.so` ./

# Start a new, final image to reduce size.
FROM debian as final

######
# TEMPORARY GREENLIGHT CLI
# THIS MAY BREAK AT ANY TIME!
RUN apt-get update && \
   apt-get install -y python3-pip wget

RUN pip install -U gl-client --break-system-packages
RUN pip install --extra-index-url=https://us-west2-python.pkg.dev/c-lightning/greenlight-pypi/simple/ -U glcli --break-system-packages
#RUN python3 -c 'import sysconfig; print(sysconfig.get_paths()["purelib"])'
RUN wget -O /usr/local/lib/python3.11/dist-packages/glcli/cli.py https://raw.githubusercontent.com/Blockstream/greenlight/2dc5a94668d41baef7275dae860c09b4a5dba198/tools/glcli/glcli/cli.py
######


ENV LD_LIBRARY_PATH=/usr/lib/libbreez
#
# # Copy the binaries and entrypoint from the builder image.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/libbreez_sdk_bindings.so /usr/lib/libbreez/
COPY --from=builder /build/main /bin/

ENTRYPOINT [ "/bin/main" ]
