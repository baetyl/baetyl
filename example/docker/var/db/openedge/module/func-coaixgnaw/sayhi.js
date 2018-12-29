#!/usr/bin/env node
// in
exports.handler = (event, context, callback) => {
    // edit event
    event.name = 'openedge';
    event.say = 'hello world';
    // get context
    event['functionName'] = context['functionName']
    event['functionInvokeID'] = context['functionInvokeID']
    event['messageQOS'] = context['messageQOS']
    event['messageTopic'] = context['messageTopic']
    // get os env
    event['USER_ID'] = process.env['USER_ID']
    // callback result
    callback(null, event);
};