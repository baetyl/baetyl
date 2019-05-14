// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('grpc');
var function_pb = require('./function_pb.js');

function serialize_openedge_FunctionMessage(arg) {
  if (!(arg instanceof function_pb.FunctionMessage)) {
    throw new Error('Expected argument of type openedge.FunctionMessage');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_openedge_FunctionMessage(buffer_arg) {
  return function_pb.FunctionMessage.deserializeBinary(new Uint8Array(buffer_arg));
}


// The function server definition.
var FunctionService = exports.FunctionService = {
  call: {
    path: '/openedge.Function/Call',
    requestStream: false,
    responseStream: false,
    requestType: function_pb.FunctionMessage,
    responseType: function_pb.FunctionMessage,
    requestSerialize: serialize_openedge_FunctionMessage,
    requestDeserialize: deserialize_openedge_FunctionMessage,
    responseSerialize: serialize_openedge_FunctionMessage,
    responseDeserialize: deserialize_openedge_FunctionMessage,
  },
};

exports.FunctionClient = grpc.makeGenericClientConstructor(FunctionService);
