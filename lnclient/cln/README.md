## Generating Go gRPC Code

The Go gRPC bindings for Core Lightning (CLN) are generated from proto files that live
in the **lightning** repository (`cln-grpc`) and in the **hold** plugin repository.

The generated Go files are written into the **hub** repository under `lnclient/cln/clngrpc` and `lnclient/cln/clngrpc_hold`.

### Prerequisites

Make sure the following tools are installed:

```bash
# protoc (>= 3.20 recommended)
protoc --version

# Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Ensure `$GOPATH/bin` is in your `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

This guide assumes the following directory structure:

```
~/dev/
├── lightning/
│   └── cln-grpc/
│       └── proto/
│           ├── node.proto
│           ├── primitives.proto
│           └── ...
├── hub/
|   └── lnclient/
|       └── cln/
|           └── clngrpc/
|           └── clngrpc_hold/
└── hold/
    └── protos/
        └── hold.proto
```

### Generating Go code

From the hub repository root, run:

```bash
protoc \
  --proto_path=../lightning/cln-grpc/proto \
  --go_out=./lnclient/cln/clngrpc \
  --go_opt=paths=source_relative \
  --go_opt=Mprimitives.proto=github.com/getAlby/hub/lnclient/cln/clngrpc \
  --go_opt=Mnode.proto=github.com/getAlby/hub/lnclient/cln/clngrpc \
  --go-grpc_out=./lnclient/cln/clngrpc \
  --go-grpc_opt=paths=source_relative \
  ../lightning/cln-grpc/proto/node.proto \
  ../lightning/cln-grpc/proto/primitives.proto
```

and if you have the hold plugin repo:

```bash
protoc \
  --proto_path=../hold/protos \
  --go_out=./lnclient/cln/clngrpc_hold \
  --go_opt=paths=source_relative \
  --go_opt=Mhold.proto=github.com/getAlby/hub/lnclient/cln/clngrpc_hold \
  --go-grpc_out=./lnclient/cln/clngrpc_hold \
  --go-grpc_opt=paths=source_relative \
  ../hold/protos/hold.proto
```
