// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('grpc');
var function_pb = require('./function_pb.js');

function serialize_baetyl_FunctionMessage(arg) {
  if (!(arg instanceof function_pb.FunctionMessage)) {
    throw new Error('Expected argument of type baetyl.FunctionMessage');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_baetyl_FunctionMessage(buffer_arg) {
  return function_pb.FunctionMessage.deserializeBinary(new Uint8Array(buffer_arg));
}


// The function server definition.
var FunctionService = exports.FunctionService = {
  call: {
    path: '/baetyl.Function/Call',
    requestStream: false,
    responseStream: false,
    requestType: function_pb.FunctionMessage,
    responseType: function_pb.FunctionMessage,
    requestSerialize: serialize_baetyl_FunctionMessage,
    requestDeserialize: deserialize_baetyl_FunctionMessage,
    responseSerialize: serialize_baetyl_FunctionMessage,
    responseDeserialize: deserialize_baetyl_FunctionMessage,
  },
};

exports.FunctionClient = grpc.makeGenericClientConstructor(FunctionService);
