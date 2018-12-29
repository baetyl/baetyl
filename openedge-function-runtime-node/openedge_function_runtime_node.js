#!/usr/bin/env node
const path = require('path');
const fs = require('fs');
const yaml = require('js-yaml');
const log4js = require('log4js');
const moment = require('moment');
const services = require('./runtime_pb_grpc.js');
const grpc = require('grpc');
const argv = require('yargs')
    .option('c', {
        alias: 'conf',
        demand: false,
        default: path.join('conf', 'conf.yml'),
        describe: 'config file path (default: conf/conf.yml)',
        type: 'string'
    })
    .usage('Usage: hello [options]')
    .help('h')
    .alias('h', 'help')
    .epilog('copyright 2018')
    .argv;

const getLogger = config => {
    if (!hasAttr(config, 'logger')) {
        return log4js.getLogger(config.name);
    }
    if (!hasAttr(config.logger, 'path')) {
        return log4js.getLogger(config.name);
    }
    let level = 'info';
    if (hasAttr(config.logger, 'level')) {
        if (config.logger.level === 'debug') {
            level = 'debug';
        }
        else if (config.logger.level === 'warn') {
            level = 'warn';
        }
        else if (config.logger.level === 'error') {
            level = 'error';
        }
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


const hasAttr = (obj, attr) => {
    if (obj instanceof Object && !(obj instanceof Array)) {
        if (obj[attr]) {
            return true;
        }
    }
    return false;
};

class ConfigError {
    constructor(message) {
        this.message = message;
        this.name = 'ConfigError';
    }
}

class NodeRuntimeModule {
    load(conf) {
        conf = conf.trim();
        if (conf && conf[0] === '{' && conf[conf.length - 1] === '}') {
            this.config = JSON.parse(conf);
        }
        else {
            const yamlData = fs.readFileSync(conf, 'utf8');
            const nativeObject = yaml.safeLoad(yamlData);
            if (nativeObject) {
                this.config = nativeObject;
            }
            else {
                this.config = {};
            }
        }
        if (!hasAttr(this.config, 'name')) {
            throw new ConfigError('Module config invalid, missing name');
        }
        if (!hasAttr(this.config, 'server')) {
            throw new ConfigError('Module config invalid, missing server');
        }
        if (!hasAttr(this.config.server, 'address')) {
            throw new ConfigError('Module config invalid, missing server address');
        }
        if (!hasAttr(this.config, 'function')) {
            throw new ConfigError('Module config invalid, missing function');
        }
        this.logger = getLogger(this.config);
        if (!hasAttr(this.config['function'], 'name')){
            throw new ConfigError('Module config invalid, missing function name');
        }
        if (!hasAttr(this.config['function'], 'handler')) {
            throw new ConfigError('Module config invalid, missing function handler');
        }
        if (!hasAttr(this.config['function'], 'codedir')) {
            throw new ConfigError('Module config invalid, missing function codedir');
        }
        const codedir = this.config['function'].codedir;
        const moduleHandler = this.config['function'].handler.split('.');
        const handlerName = moduleHandler[1];
        const moduleName = require(path.join(process.cwd(), codedir, moduleHandler[0]));
        const functionHandle = moduleName[handlerName];
        let maxMessageSize = 4 * 1024 * 1024;
        if (hasAttr(this.config.server, 'message')
            && hasAttr(this.config.server.message, 'length')
            && hasAttr(this.config.server.message.length, 'max')) {
            maxMessageSize = this.config.server.message.length.max;
        }
        this.server = new grpc.Server({
            'grpc.max_send_message_length': maxMessageSize,
            'grpc.max_receive_message_length': maxMessageSize
        });
        this.server.addService(services.RuntimeService, {
            handle: (call, callback) => (this.handle(functionHandle, call, callback))
        });
        this.server.bind(this.config.server.address, grpc.ServerCredentials.createInsecure());
    }

    start() {
        this.server.start();
        this.logger.info('module starting');
    }

    close(callback) {
        const timeout = new Number(this.config.server.timeout / 1e6);
        setTimeout(() => {
            this.server.forceShutdown();
            this.logger.info('module closed');
            callback();
        }, timeout);
    }

    handle(functionHandle, call, callback) {
        const ctx = {};
        ctx.messageQOS = call.request.getQos();
        ctx.messageTopic = call.request.getTopic();
        ctx.functionName = call.request.getFunctionname();
        ctx.functionInvokeID = call.request.getFunctioninvokeid();
        ctx.invokeid = call.request.getFunctioninvokeid();
        let msg = {};
        const Payload = call.request.getPayload();
        if (Payload) {
            try {
                const payloadStr = Buffer.from(Payload).toString();
                msg = JSON.parse(payloadStr);
            }
            catch (error) {
                msg = Payload; // raw data, not json format
            }
        }
        functionHandle(
            msg,
            ctx,
            (err, respMsg) => {
                if (err != null) {
                    throw new Error(err);
                }
                if (!respMsg) {
                    call.request.setPayload([]);
                }
                else if (respMsg instanceof Object && !(respMsg instanceof Array)) {
                    call.request.setPayload(Buffer.from(JSON.stringify(respMsg)));
                }
                else {
                    call.request.setPayload(Buffer.from(respMsg));
                }
                callback(null, call.request);
            }
        );
    }
}

function main() {
    const runtimeModule = new NodeRuntimeModule();
    runtimeModule.load(argv.conf);
    runtimeModule.start();

    function closeServer() {
        runtimeModule.close(() => log4js.shutdown(() => process.exit(0)));
    }

    process.on('SIGINT', () => {
        closeServer();
    });
    process.on('SIGTERM', () => {
        closeServer();
    });
}

main();
