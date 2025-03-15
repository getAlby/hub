#!/usr/bin/env bash

ARCH="$1"

if [[ "$ARCH" == "amd64" ]]; then
    cp "$(go list -m -f "{{.Dir}}" github.com/getAlby/ldk-node-go)"/ldk_node/x86_64-unknown-linux-gnu/libldk_node.so ./
elif [[ "$ARCH" == "arm64" ]]; then
    cp "$(go list -m -f "{{.Dir}}" github.com/getAlby/ldk-node-go)"/ldk_node/aarch64-unknown-linux-gnu/libldk_node.so ./
else
    echo "Invalid ARCH value"
    exit 1
fi
