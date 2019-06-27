#!/usr/bin/env node
const path = require('path');
const fs = require('fs');
const log4js = require('log4js');
const moment = require('moment');
const services = require('./function_grpc_pb.js');
const grpc = require('grpc');
const YAML = require('yaml');
const util = require('util');
const errFormat = 'Exception calling application: %s'
const argv = require('yargs')
    .option('c', {
        alias: 'conf',
        demand: false,
        default: path.join('etc', 'openedge', 'service.yml'),
        describe: 'config file path',
        type: 'string'
    })
    .argv;

const hasAttr = (obj, attr) => {
    if (obj instanceof Object && !(obj instanceof Array)) {
        if (obj[attr] != undefined) {
            return true;
        }
    }
    return false;
};

const getLogger = config => {
    if (!hasAttr(config, 'logger')) {
        return log4js.getLogger(config.name);
    }
    if (!hasAttr(config.logger, 'path')) {
        return log4js.getLogger(config.name);
    }
    let level = 'info';
    if (hasAttr(config.logger, 'level')) {
        level = config.logger.level;
    }

    let backupCount = 15;
    if (hasAttr(config.logger, 'backupCount') && hasAttr(config.logger.backupCount, 'max')) {
        backupCount = config.logger.backupCount.max;
    }
    log4js.addLayout('openedge', () => logEvent => {
        const asctime = moment(logEvent.startTime).format('YYYY-MM-DD HH:mm:ss');
        const name = logEvent.categoryName;
        const levelname = logEvent.level.levelStr;
        const message = logEvent.data;
        return `${asctime} - ${name} - ${levelname} - ${message}`;
    });
    log4js.configure({
        appenders: {
            file: {
                type: 'file',
                filename: config.logger.path,
                layout: {type: 'openedge'},
                backups: backupCount,
                compress: true,
                encoding: 'utf-8'
            }
        },
        categories: {
            default: {appenders: ['file'], level}
        }
    });
    const logger = log4js.getLogger(config.name);
    return logger;
};

const getFunctions = functions => {

    functionsHandle = {};

    functions.forEach(function(ele){
        
        if (ele.name == undefined || ele.handler == undefined || ele.codedir == undefined){
            throw new Error('config invalid, missing function name, handler or codedir');
        }
        
        const codedir = ele.codedir;
        const moduleHandler = ele.handler.split('.');
        const handlerName = moduleHandler[1];
        const moduleName = require(path.join(process.cwd(), codedir, moduleHandler[0]));
        const functionHandle = moduleName[handlerName];
        functionsHandle[ele.name] = functionHandle;
    });
    
    return functionsHandle;
};

const getGrpcServer = config => {

    let maxMessageSize = 4 * 1024 * 1024;
    if (hasAttr(config['server'], 'message')
        && hasAttr(config['server']['message'], 'length')
        && hasAttr(config['server']['message']['length'], 'max')) {
        maxMessageSize = config['server']['message']['length']['max'];
    }
    let server = new grpc.Server({
        'grpc.max_send_message_length': maxMessageSize,
        'grpc.max_receive_message_length': maxMessageSize
    });

    let credentials = undefined;
    
    if (hasAttr(config.server, 'ca') 
        && hasAttr(config.server, 'key') 
        && hasAttr(config.server, 'cert')) {
            
        credentials = grpc.ServerCredentials.createSsl(
            fs.readFileSync(config['server']['ca']), [{
            cert_chain: fs.readFileSync(config['server']['cert']),
            private_key: fs.readFileSync(config['server']['key'])
        }], true);
    }else {
        credentials = grpc.ServerCredentials.createInsecure();
    }

    server.bind(config['server']['address'], credentials);
    return server;
}

class NodeRuntimeModule {
    Load(confpath) {
        this.config = YAML.parse(fs.readFileSync(confpath).toString()); 

        if(hasAttr(process.env, 'OPENEDGE_SERVICE_INSTANCE_NAME')){
            this.config['name'] = process.env['OPENEDGE_SERVICE_INSTANCE_NAME']
        }else if(hasAttr(process.env, 'OPENEDGE_SERVICE_NAME')){
            this.config['name'] = process.env['OPENEDGE_SERVICE_NAME']
        }

        if(hasAttr(process.env, 'OPENEDGE_SERVICE_INSTANCE_ADDRESS')){
            if(!hasAttr(this.config, 'server')){
                this.config['server'] = {}
            }
            this.config['server']['address'] = process.env['OPENEDGE_SERVICE_INSTANCE_ADDRESS'];
        }else if(hasAttr(process.env, 'OPENEDGE_SERVICE_NAME')){
            if(!hasAttr(this.config, 'server')){
                this.config['server'] = {}
            }
            this.config['server']['address'] = process.env['OPENEDGE_SERVICE_NAME'];
        }

        if (!hasAttr(this.config, 'name')) {
            throw new Error('Module config invalid, missing name');
        }
        if (!hasAttr(this.config, 'server')) {
            throw new Error('Module config invalid, missing server');
        }
        if (!hasAttr(this.config.server, 'address')) {
            throw new Error('Module config invalid, missing server address');
        }
        if (!hasAttr(this.config, 'functions')) {
            throw new Error('Module config invalid, missing functions');
        }
        
        this.logger = getLogger(this.config);
        const functionsHandle = getFunctions(this.config['functions']);
        this.server = getGrpcServer(this.config);
        
        this.server.addService(services.FunctionService, {
            call: (call, callback) => (this.Call(functionsHandle, call, callback))
        });
    }
    Start() {
        this.server.start();
        this.logger.info('service starting');
    }
    Close(callback) {
        if (hasAttr(this.config.server, 'timeout')) {
            const timeout = new Number(this.config.server.timeout / 1e9);
            setTimeout(() => {
                this.server.forceShutdown();
                this.logger.info('service closed');
                callback();
            }, timeout);
        }else{
            this.server.forceShutdown();
            this.logger.info('service closed');
        }
    }
    Call(functionsHandle, call, callback) {

        const ctx = {};
        ctx.messageQOS = call.request.getQos();
        ctx.messageTopic = call.request.getTopic();
        ctx.functionName = call.request.getFunctionname();
        ctx.messageTimestamp = call.request.getTimestamp();
        ctx.functionInvokeID = call.request.getFunctioninvokeid();
        ctx.invokeid = call.request.getFunctioninvokeid();

        let msg = Buffer.from([]);
        const Payload = call.request.getPayload();

        if (Payload) {
            try {
                const payloadString = Buffer.from(Payload).toString();
                msg = JSON.parse(payloadString);
            }
            catch (error) {
                msg = Buffer.from(Payload); // raw data, not json format
            }
        }else{
            msg = {};
        }

        if (functionsHandle[ctx.functionName] == undefined){
            return callback(new Error(util.format(errFormat, "function not found")));
        }

        let functionHandle = functionsHandle[ctx.functionName];
        functionHandle(
            msg,
            ctx,
            (err, respMsg) => {
                if (err != null) {
                    err = util.format('(\'[UserCodeInvoke]\', %s)', err.toString())
                    return callback(new Error(util.format(errFormat, err)));
                }

                if (respMsg == "" || respMsg == undefined || respMsg == null){
                    call.request.setPayload("");
                }else if (Buffer.isBuffer(respMsg)) {
                    call.request.setPayload(respMsg);
                }
                else {
                    try {
                        const jsonString = JSON.stringify(respMsg);
                        call.request.setPayload(Buffer.from(jsonString));
                    }
                    catch (error) {
                        err = util.format('(\'[UserCodeReturn]\', %s)', error.toString())
                        return callback(new Error(util.format(errFormat, err)));
                    }
                }
                callback(null, call.request);
            }
        );
    }
}

(() => {
    const runtimeModule = new NodeRuntimeModule();
    runtimeModule.Load(argv.c);
    runtimeModule.Start();
    function closeServer() {
        runtimeModule.Close(() => log4js.shutdown(() => process.exit(0)));
    }
    process.on('SIGINT', () => {
        closeServer();
    });
    process.on('SIGTERM', () => {
        closeServer();
    });
})();
