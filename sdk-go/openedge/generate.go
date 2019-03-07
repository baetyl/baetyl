package openedge

//go:generate echo "If protoc not installed, please get it from https://github.com/protocolbuffers/protobuf/releases"
//go:generate protoc -I. --go_out=plugins=grpc:. function.proto
//go:generate python -m grpc_tools.protoc -I. --python_out=../../openedge-function-python27 --grpc_python_out=../../openedge-function-python27 function.proto
