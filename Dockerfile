FROM node:18-alpine as frontend
WORKDIR /build
COPY frontend ./frontend
RUN cd frontend && yarn install && yarn build

FROM golang:1.20-alpine as builder

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

# Build the application
RUN go build -o main

# Start a new, final image to reduce size.
FROM alpine as final

# Copy the binaries and entrypoint from the builder image.
COPY --from=builder /build/main /bin/

ENTRYPOINT [ "/bin/main" ]
