package openedge

//go:generate echo "If protoc not installed, please get it from https://github.com/protocolbuffers/protobuf/releases"
//go:generate protoc -I. --go_out=plugins=grpc:. function.proto
//go:generate python3 -m grpc_tools.protoc -I. --python_out=../../openedge-function-python --grpc_python_out=../../openedge-function-python function.proto
