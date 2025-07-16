#!/usr/bin/env bash

set -eo pipefail

generate_protos() {
  package="$1"
  proto_dirs=$(find $package -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
  for dir in $proto_dirs; do
    for file in $(find "${dir}" -maxdepth 1 -name '*.proto'); do
      if grep go_package "$file" &>/dev/null; then
        buf generate --template buf.gen.gogo.yaml "$file"
      fi
    done
  done
}

echo "Generating gogo proto code"
cd ./proto

# generate_protos "./stride"
# generate_protos "./dydxprotocol"
# generate_protos "./neutron"
generate_protos "./umee"

cd ..

# move proto files to the right places
#
# Note: Proto files are suffixed with the current binary version.
mkdir -p ./pkg/proto
# cp -r github.com/Stride-Labs/stride/v26/x/* ./pkg/proto/
# cp -r github.com/neutron-org/neutron/v5/x/* ./pkg/proto/
cp -r github.com/umee-network/umee/v6/x/* ./pkg/proto/umee/
# cp -r github.com/dydxprotocol/v4-chain/protocol/x/* ./pkg/proto/

rm -rf github.com
