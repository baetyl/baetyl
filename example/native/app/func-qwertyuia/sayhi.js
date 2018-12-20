exports.handler = (event, context, callback) => {
    event.name = 'openedge';
    event.say = 'hi';
    callback(null, event);
};