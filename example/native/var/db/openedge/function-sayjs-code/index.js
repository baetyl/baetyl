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
    if (hasAttr(context, 'functionName')) {
        event['functionName'] = context['functionName'];
    }
    if (hasAttr(context, 'functionInvokeID')) {
        event['functionInvokeID'] = context['functionInvokeID'];
    }
    if (hasAttr(context, 'messageQOS')) {
        event['messageQOS'] = context['messageQOS'];
    }
    if (hasAttr(context, 'messageTopic')) {
        event['messageTopic'] = context['messageTopic'];
    }
};

exports.handler = (event, context, callback) => {
    // support Buffer & json object
    if (Buffer.isBuffer(event)) {
        const message = event.toString();
        event = {};

        passParameters(event, context);
        event['node'] = 'hello world';
    }
    else {

        passParameters(event, context);
        event['node'] = 'hello world';
    }
    // callback result
    callback(null, event);
};