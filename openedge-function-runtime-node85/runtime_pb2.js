var jspb = require('google-protobuf');
var goog = jspb;
var global = Function('return this')();

goog.exportSymbol('proto.runtime.Message', null, global);
goog.exportSymbol('proto.runtime.Runtime', null, global);

proto.runtime.Message = function (opt_data) {
    jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.runtime.Message, jspb.Message);
if (goog.DEBUG && !COMPILED) {
    proto.runtime.Message.displayName = 'proto.runtime.Message';
}


if (jspb.Message.GENERATE_TO_OBJECT) {

    proto.runtime.Message.prototype.toObject = function (opt_includeInstance) {
        return proto.runtime.Message.toObject(opt_includeInstance, this);
    };

    proto.runtime.Message.toObject = function (includeInstance, msg) {
        var f, obj = {
            QOS: msg.getQOS(),
            Topic: msg.getTopic(),
            Payload: msg.getPayload(),
            FunctionName: msg.getFunctionName(),
            FunctionInvokeID: msg.getFunctionInvokeID()
        };

        if (includeInstance) {
            obj.$jspbMessageInstance = msg;
        }
        return obj;
    };
}

proto.runtime.Message.deserializeBinary = function (bytes) {
    var reader = new jspb.BinaryReader(bytes);
    var msg = new proto.runtime.Message;
    return proto.runtime.Message.deserializeBinaryFromReader(msg, reader);
};

proto.runtime.Message.deserializeBinaryFromReader = function (msg, reader) {
    while (reader.nextField()) {
        if (reader.isEndGroup()) {
            break;
        }
        var field = reader.getFieldNumber();
        switch (field) {
            case 1:
                var value = /** @type {number} */ (reader.readInt32());
                msg.setQOS(value);
                break;
            case 2:
                var value = /** @type {string} */ (reader.readString());
                msg.setTopic(value);
                break;
            case 3:
                var value = /** @type {byte} */ (reader.readBytes());
                msg.setPayload(value);
                break;
            case 4:
                var value = /** @type {string} */ (reader.readString());
                msg.setFunctionName(value);
                break;
            case 5:
                var value = /** @type {string} */ (reader.readString());
                msg.setFunctionInvokeID(value);
                break;
            default:
                reader.skipField();
                break;
        }
    }
    return msg;
};

proto.runtime.Message.serializeBinaryToWriter = function (message, writer) {
    message.serializeBinaryToWriter(writer);
};

proto.runtime.Message.prototype.serializeBinary = function () {
    var writer = new jspb.BinaryWriter();
    this.serializeBinaryToWriter(writer);
    return writer.getResultBuffer();
};

proto.runtime.Message.prototype.serializeBinaryToWriter = function (writer) {
    var f = undefined;
    f = this.getQOS();
    if (f !== 0) {
        writer.writeInt32(
            1,
            f
        );
    }
    f = this.getTopic();
    if (f.length > 0) {
        writer.writeString(
            2,
            f
        );
    }
    f = this.getPayload();
    if (f.length > 0) {
        writer.writeBytes(
            3,
            f
        );
    }
    f = this.getFunctionName();
    if (f.length > 0) {
        writer.writeString(
            4,
            f
        );
    }
    f = this.getFunctionInvokeID();
    if (f.length > 0) {
        writer.writeString(
            5,
            f
        );
    }
};

proto.runtime.Message.prototype.cloneMessage = function () {
    return /** @type {!proto.runtime.Message} */ (jspb.Message.cloneMessage(this));
};

proto.runtime.Message.prototype.getQOS = function () {
    return /** @type {number} */ (jspb.Message.getFieldProto3(this, 1, 0));
};

proto.runtime.Message.prototype.setQOS = function (value) {
    jspb.Message.setField(this, 1, value);
};

proto.runtime.Message.prototype.getTopic = function () {
    return /** @type {string} */ (jspb.Message.getFieldProto3(this, 2, ""));
};

proto.runtime.Message.prototype.setTopic = function (value) {
    jspb.Message.setField(this, 2, value);
};

proto.runtime.Message.prototype.getPayload = function () {
    return /** @type {byte} */ (jspb.Message.getFieldProto3(this, 3, []));
};

proto.runtime.Message.prototype.setPayload = function (value) {
    jspb.Message.setField(this, 3, value);
};

proto.runtime.Message.prototype.getFunctionName = function () {
    return /** @type {string} */ (jspb.Message.getFieldProto3(this, 4, ""));
};
proto.runtime.Message.prototype.setFunctionName = function (value) {
    jspb.Message.setField(this, 4, value);
};

proto.runtime.Message.prototype.getFunctionInvokeID = function () {
    return /** @type {string} */ (jspb.Message.getFieldProto3(this, 5, ""));
};
proto.runtime.Message.prototype.setFunctionInvokeID = function (value) {
    jspb.Message.setField(this, 5, value);
};

proto.runtime.Runtime = function (opt_data) {
    jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.runtime.Runtime, jspb.Message);
if (goog.DEBUG && !COMPILED) {
    proto.runtime.Runtime.displayName = 'proto.runtime.Runtime';
}


if (jspb.Message.GENERATE_TO_OBJECT) {

    proto.runtime.Runtime.prototype.toObject = function (opt_includeInstance) {
        return proto.runtime.Runtime.toObject(opt_includeInstance, this);
    };

    proto.runtime.Runtime.toObject = function (includeInstance, msg) {
        var f, obj = {
            Handle: (f = msg.getHandle()) && proto.runtime.Message.toObject(includeInstance, f)
        };

        if (includeInstance) {
            obj.$jspbMessageInstance = msg;
        }
        return obj;
    };
}

proto.runtime.Runtime.deserializeBinary = function (bytes) {
    var reader = new jspb.BinaryReader(bytes);
    var msg = new proto.runtime.Runtime;
    return proto.runtime.Runtime.deserializeBinaryFromReader(msg, reader);
};

proto.runtime.Runtime.deserializeBinaryFromReader = function (msg, reader) {
    while (reader.nextField()) {
        if (reader.isEndGroup()) {
            break;
        }
        var field = reader.getFieldNumber();
        switch (field) {
            case 1:
                var value = new proto.runtime.Message;
                reader.readMessage(value, proto.runtime.Message.deserializeBinaryFromReader);
                msg.setHandle(value);
                break;
            default:
                reader.skipField();
                break;
        }
    }
    return msg;
};

proto.runtime.Runtime.serializeBinaryToWriter = function (message, writer) {
    message.serializeBinaryToWriter(writer);
};

proto.runtime.Runtime.prototype.serializeBinary = function () {
    var writer = new jspb.BinaryWriter();
    this.serializeBinaryToWriter(writer);
    return writer.getResultBuffer();
};

proto.runtime.Runtime.prototype.serializeBinaryToWriter = function (writer) {
    var f = undefined;
    f = this.getHandle();
    if (f != null) {
        writer.writeMessage(
            1,
            f,
            proto.runtime.Message.serializeBinaryToWriter
        );
    }
};

proto.runtime.Runtime.prototype.cloneMessage = function () {
    return /** @type {!proto.runtime.Runtime} */ (jspb.Message.cloneMessage(this));
};

proto.runtime.Runtime.prototype.getHandle = function () {
    return /** @type{proto.runtime.Message} */ (
        jspb.Message.getWrapperField(this, proto.runtime.Message, 1));
};


/** @param {proto.runtime.Message|undefined} value  */
proto.runtime.Runtime.prototype.setHandle = function (value) {
    jspb.Message.setWrapperField(this, 1, value);
};


proto.runtime.Runtime.prototype.clearHandle = function () {
    this.setHandle(undefined);
};

goog.object.extend(exports, proto.runtime);