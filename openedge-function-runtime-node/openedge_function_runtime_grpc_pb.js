// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('grpc');
var openedge_function_runtime_pb = require('./openedge_function_runtime_pb.js');

function serialize_runtime_Message(arg) {
  if (!(arg instanceof openedge_function_runtime_pb.Message)) {
    throw new Error('Expected argument of type runtime.Message');
  }
  return new Buffer(arg.serializeBinary());
}

function deserialize_runtime_Message(buffer_arg) {
  return openedge_function_runtime_pb.Message.deserializeBinary(new Uint8Array(buffer_arg));
}


// The runtime definition.
var RuntimeService = exports.RuntimeService = {
  // Handle handles request
  handle: {
    path: '/runtime.Runtime/Handle',
    requestStream: false,
    responseStream: false,
    requestType: openedge_function_runtime_pb.Message,
    responseType: openedge_function_runtime_pb.Message,
    requestSerialize: serialize_runtime_Message,
    requestDeserialize: deserialize_runtime_Message,
    responseSerialize: serialize_runtime_Message,
    responseDeserialize: deserialize_runtime_Message,
  },
};

exports.RuntimeClient = grpc.makeGenericClientConstructor(RuntimeService);
