#!/usr/bin/env node
const path = require('path');
const fs = require('fs');
const yaml = require('js-yaml');
const { getLogger, hasAttr, ConfigError } = require('./utils.js');
const services = require('./runtime_pb2_grpc.js');
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

const handle = (self, call, callback) => {
    const ctx = {}
    ctx['messageQOS'] = call.request.getQOS();
    ctx['messageTopic'] = call.request.getTopic();
    ctx['functionName'] = call.request.getFunctionName();
    ctx['functionInvokeID'] = call.request.getFunctionInvokeID();
    ctx['invokeid'] = call.request.getFunctionInvokeID();
    let msg = {};
    const Payload = call.request.getPayload();
    if (Payload) {
        try {
            const payloadStr = new Buffer(Payload).toString();
            msg = JSON.parse(payloadStr);
        } catch (error) {
            msg = Payload; // raw data, not json format
        }
    }
    self.functionHandle(
        msg,
        ctx,
        (err, respMsg) => {
            if (err != null) {
                throw new Error(err);
            }
            if (!respMsg) {
                call.request.setPayload([]);
            } else {
                if (respMsg instanceof Object && !(respMsg instanceof Array)) {
                    call.request.setPayload(new Buffer(JSON.stringify(respMsg)));
                } else {
                    call.request.setPayload(new Buffer(respMsg));
                }
        
            }
            callback(null, call.request);
        }
    );  
};

const start = self => {
    self.logger.info("module starting");
    self.server.start();
}

const close = self => {
    self.server.forceShutdown();
    self.logger.info("module closed");
}

class NodeRuntimeModule {
    load(conf) {
        conf = conf.trim()
        if (conf && conf[0] === '{' && conf[conf.length - 1] === '}') {
            this.config = JSON.parse(conf);
        } else {
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
        if (!hasAttr(this.config, 'function')) {
            throw new ConfigError('Module config invalid, missing function');
        }
        this.logger = getLogger(this.config);
        if (!hasAttr(this.config.function, 'name')
            || !hasAttr(this.config.function, 'handler')
            || !hasAttr(this.config.function, 'codedir')) {
            throw new ConfigError('Module config invalid, missing function name, handler or codedir');
        }
        const codedir = this.config['function']['codedir'];
        const moduleHandler = this.config['function']['handler'].split('.')
        const handlerName = moduleHandler[1];
        const moduleName = require(path.join(process.cwd(), codedir, moduleHandler[0]));
        this.functionHandle = moduleName[handlerName];
        let maxMessageSize = 4 * 1024 * 1024;
        if (hasAttr(this.config.server, 'message')
            && hasAttr(this.config.server.message, 'length')
            && hasAttr(this.config.server.message.length, 'max')) {
            maxMessageSize = this.config['server']['message']['length']['max'];
        }
        if (!hasAttr(this.config.server, 'address')) {
            throw new ConfigError('Module config invalid, missing server address');
        }
        this.server = new grpc.Server({ 'grpc.max_send_message_length': maxMessageSize, 'grpc.max_receive_message_length': maxMessageSize });
        this.server.addService(services.RuntimeService, {
            handle: (call, callback) => (handle(this, call, callback)),
            start: () => (start(this)),
            stop: () => (stop(this)),
        });
        this.server.bind(this.config.server.address, grpc.ServerCredentials.createInsecure());
    }
}

async function main() {
    const nodeModule = new NodeRuntimeModule();
    nodeModule.load(argv.conf);
    start(nodeModule);

    function closeServer() {
        close(nodeModule);
        process.exit(0);
    }
    process.on('SIGINT', function () {
        closeServer();
    });
    process.on('SIGTERM', function () {
        closeServer();
    });
}

main();