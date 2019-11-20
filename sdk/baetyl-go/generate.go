package baetyl

// https://github.com/protocolbuffers/protobuf/releases"
// https://www.npmjs.com/package/protoc-gen-grpc

//go:generate protoc -I. --go_out=plugins=grpc:. function.proto
//go:generate python3 -m grpc_tools.protoc -I. --python_out=../../baetyl-function-python3 --grpc_python_out=../../baetyl-function-python3 function.proto
//go:generate python3 -m grpc_tools.protoc -I. --python_out=../../baetyl-function-python2 --grpc_python_out=../../baetyl-function-python2 function.proto
//go:generate protoc-gen-grpc -I=. --js_out=import_style=commonjs,binary:../../baetyl-function-node8 --grpc_out=../../baetyl-function-node8 function.proto
//go:generate protoc -I. --go_out=plugins=grpc:. kv.proto

//go:generate ./templates/gen.sh agent
//go:generate ./templates/gen.sh hub
//go:generate ./templates/gen.sh remote-mqtt
//go:generate ./templates/gen.sh timer
//go:generate ./templates/gen.sh function-manager
//go:generate ./templates/gen.sh function-node8
//go:generate ./templates/gen.sh function-python2
//go:generate ./templates/gen.sh function-python3
