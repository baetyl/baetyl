#!/usr/bin/env node

const hasAttr = (obj, attr) => {
    if (obj instanceof Object && !(obj instanceof Array)) {
        if (obj[attr] != undefined) {
            return true;
        }
    }
    return false;
};

const passParameters = (event, context) => {
    if (hasAttr(context, 'messageQOS')) {
        event['messageQOS'] = context['messageQOS'];
    }
    if (hasAttr(context, 'messageTopic')) {
        event['messageTopic'] = context['messageTopic'];
    }
    if (hasAttr(context, 'messageTimestamp')) {
        event['messageTimestamp'] = context['messageTimestamp'];
    }
    if (hasAttr(context, 'functionName')) {
        event['functionName'] = context['functionName'];
    }
    if (hasAttr(context, 'functionInvokeID')) {
        event['functionInvokeID'] = context['functionInvokeID'];
    }
};

exports.handler = (event, context, callback) => {
    // support Buffer & json object
    if (Buffer.isBuffer(event)) {
        const message = event.toString();
        event = {}
        event["bytes"] = message;
    }
    else if("err" in event) {
        throw new TypeError(event["err"])
    }

    passParameters(event, context);
    event['Say'] = 'Hello OpenEdge'
    // callback result
    callback(null, event);
};