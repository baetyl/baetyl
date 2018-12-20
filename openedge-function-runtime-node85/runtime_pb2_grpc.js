'use strict';
var grpc = require('grpc');
var runtime_pb = require('./runtime_pb2.js');

function serialize_Message(arg) {
    if (!(arg instanceof runtime_pb.Message)) {
        throw new Error('Expected argument of type Message');
    }
    return new Buffer(arg.serializeBinary());
}

function deserialize_Message(buffer_arg) {
    return runtime_pb.Message.deserializeBinary(new Uint8Array(buffer_arg));
}

// Interface exported by the server.
var RuntimeService = exports.RuntimeService = {
    handle: {
        path: '/runtime.Runtime/Handle',
        requestStream: false,
        responseStream: false,
        requestType: runtime_pb.Message,
        responseType: runtime_pb.Message,
        requestDeserialize: deserialize_Message,
        responseSerialize: serialize_Message
    }
};

exports.RuntimeClient = grpc.makeGenericClientConstructor(RuntimeService);