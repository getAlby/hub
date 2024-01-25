FROM node:18-alpine as frontend
WORKDIR /build
COPY frontend ./frontend
RUN cd frontend && yarn install && yarn build

FROM golang:latest as builder

RUN apt-get update && \
    apt-get install -y gcc

ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

# Move to working directory /build
WORKDIR /build

# Copy and download dependency using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container
COPY . .

# Copy frontend dist files into the container
COPY --from=frontend /build/frontend/dist ./frontend/dist

RUN go build -o main .

# Start a new, final image to reduce size.
FROM alpine as final

# FROM gcr.io/distroless/static-debian11

# USER small-user:small-user

# Copy the binaries and entrypoint from the builder image.
COPY --from=builder /build/main /bin/

ENTRYPOINT [ "/bin/main" ]
