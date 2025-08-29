#!/usr/bin/env bash

WORKDIR=`dirname $(realpath $0)`
cd $WORKDIR

# pushd ./acl
cargo build --target=wasm32-unknown-unknown --release
cp target/wasm32-unknown-unknown/release/user_logging.wasm /tmp
# popd

