#!/usr/bin/env bash

ARCH="$1"

if [[ "$ARCH" == "amd64" ]]; then
    cp "$(go list -m -f "{{.Dir}}" github.com/breez/breez-sdk-go)"/breez_sdk/lib/linux-amd64/libbreez_sdk_bindings.so ./
    cp "$(go list -m -f "{{.Dir}}" github.com/getAlby/glalby-go)"/glalby/x86_64-unknown-linux-gnu/libglalby_bindings.so ./
    cp "$(go list -m -f "{{.Dir}}" github.com/getAlby/ldk-node-go)"/ldk_node/x86_64-unknown-linux-gnu/libldk_node.so ./
elif [[ "$ARCH" == "arm64" ]]; then
    cp "$(go list -m -f "{{.Dir}}" github.com/breez/breez-sdk-go)"/breez_sdk/lib/linux-aarch64/libbreez_sdk_bindings.so ./
    cp "$(go list -m -f "{{.Dir}}" github.com/getAlby/glalby-go)"/glalby/aarch64-unknown-linux-gnu/libglalby_bindings.so ./
    cp "$(go list -m -f "{{.Dir}}" github.com/getAlby/ldk-node-go)"/ldk_node/aarch64-unknown-linux-gnu/libldk_node.so ./
else
    echo "Invalid ARCH value"
    exit 1
fi
