#!/usr/bin/env bash

[[ "$debug" ]] && set -x

# protoc -I=. -I="$GOPATH/pkg/mod"  --gogofaster_out=. common.proto

cd "$(dirname "$0")" || exit 1
cd ../.. || exit 1
ROOT=$(pwd)
cd - || exit 1

for f in `find ./ -maxdepth 1 -name "*.go" -o -name "*.sql" -o -name "*.json"`; do
    echo "cleaning $f..."
    rm -f "$f"
done

protoc -I=. -I="$GOPATH/src" -I="$ROOT" "common.proto" \
  --gofast_out=Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:.
[[ $? -ne 0 ]] && exit 1

protoc -I=. -I="$GOPATH/src" -I="$ROOT" "manager.proto" \
  --gofast_out=plugins=grpc,Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:.
[[ $? -ne 0 ]] && exit 1

protoc -I=. -I="$GOPATH/src" -I="$ROOT" "worker.proto" \
  --descriptor_set_out=worker.protoset \
  --include_imports \
  --gofast_out=plugins=grpc,Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:.
[[ $? -ne 0 ]] && exit 1

if [[ "$FILE_NAME" == "grpc.proto" ]]; then
    protoc -I=. -I="$GOPATH/src" "$FILE_NAME" -I="$ZPLUS_GO_ROOT" \
      --gofast_out=plugins=grpc,Mcommon.proto=module/agent:. \
      --agent-grpc_out=Mcommon.proto=module/agent:.
    [[ $? -ne 0 ]] && exit 1
else
  protoc -I=. -I="$GOPATH/src" -I="$PROJECT_ROOT/proto"  -I="$ZPLUS_PROTO_ROOT" "$FILE_NAME" \
    --gogo_out=Mmanager.proto=module/manager,Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:.\
    --custom_out=Mmanager.proto=module/manager:.
  [[ $? -ne 0 ]] && exit 1

  #perl -pi -e 's/(json:"[a-zA-Z0-9_]+),omitempty/$1/g' "${FILE_NAME%%.*}.pb.go"
  easyjson -all "${FILE_NAME%%.*}.pb.go"
  [[ $? -ne 0 ]] && exit 1
fi
