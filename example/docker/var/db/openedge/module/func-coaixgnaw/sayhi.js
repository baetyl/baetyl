#!/usr/bin/env node

exports.handler = (event, context, callback) => {
    // support Buffer & json object
    if (Buffer.isBuffer(event)) {
       const message = event.toString();
       event = {};
       event['type'] = 'Buffer';
       event['name'] = 'openedge';
       event['say'] = message;
    }
    else {
        event['type'] = 'json';
        event['name'] = 'openedge';
        event['say'] = 'hello world';
    }
    // get context
    event['functionName'] = context['functionName']
    // event['functionInvokeID'] = context['functionInvokeID']
    // event['messageQOS'] = context['messageQOS']
    // event['messageTopic'] = context['messageTopic']
    // get os env
    event['USER_ID'] = process.env['USER_ID']
    // callback result
    callback(null, event);
};