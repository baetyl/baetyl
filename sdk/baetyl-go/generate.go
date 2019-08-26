package baetyl

// https://github.com/protocolbuffers/protobuf/releases"
// https://www.npmjs.com/package/protoc-gen-grpc

//go:generate protoc -I. --go_out=plugins=grpc:. function.proto
//go:generate python3 -m grpc_tools.protoc -I. --python_out=../../baetyl-function-python --grpc_python_out=../../baetyl-function-python function.proto
//go:generate protoc-gen-grpc -I=. --js_out=import_style=commonjs,binary:../../baetyl-function-node --grpc_out=../../baetyl-function-node function.proto
