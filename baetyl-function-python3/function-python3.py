#!/usr/bin/env python3
# -*- coding:utf-8 -*-
"""
grpc server for python3 function
"""

import argparse
import importlib
import os
import sys
import time
import grpc
import yaml
import json
import signal
from concurrent import futures
import function_pb2
import function_pb2_grpc
import logging
import logging.handlers

_ONE_DAY_IN_SECONDS = 60 * 60 * 24


class mo(function_pb2_grpc.FunctionServicer):
    """
    grpc server module for python3 function
    """

    def Load(self, conf):
        """
        load config and init module
        """
        self.config = yaml.load(open(conf, 'r').read(), Loader=yaml.FullLoader)

        # overwrite config from env
        if 'BAETYL_SERVICE_INSTANCE_NAME' in os.environ:
            self.config['name'] = os.environ['BAETYL_SERVICE_INSTANCE_NAME']
        elif 'OPENEDGE_SERVICE_INSTANCE_NAME' in os.environ:
            self.config['name'] = os.environ['OPENEDGE_SERVICE_INSTANCE_NAME']
        elif 'OPENEDGE_SERVICE_NAME' in os.environ:  # deprecated
            self.config['name'] = os.environ['OPENEDGE_SERVICE_NAME']

        if 'BAETYL_SERVICE_INSTANCE_ADDRESS' in os.environ:
            if 'server' not in self.config:
                self.config['server'] = {}
            self.config['server']['address'] = os.environ['BAETYL_SERVICE_INSTANCE_ADDRESS']
        elif 'OPENEDGE_SERVICE_INSTANCE_ADDRESS' in os.environ:
            if 'server' not in self.config:
                self.config['server'] = {}
            self.config['server']['address'] = os.environ['OPENEDGE_SERVICE_INSTANCE_ADDRESS']
        elif 'OPENEDGE_SERVICE_ADDRESS' in os.environ:  # deprecated
            if 'server' not in self.config:
                self.config['server'] = {}
            self.config['server']['address'] = os.environ['OPENEDGE_SERVICE_ADDRESS']

        if 'name' not in self.config:
            raise Exception('config invalid, missing name')
        if 'server' not in self.config:
            raise Exception('config invalid, missing server')
        if 'address' not in self.config['server']:
            raise Exception('config invalid, missing server address')
        if 'functions' not in self.config:
            raise Exception('config invalid, missing functions')

        self.log = get_logger(self.config)
        self.functions = get_functions(self.config['functions'])
        self.server = get_grpc_server(self.config['server'])
        function_pb2_grpc.add_FunctionServicer_to_server(self, self.server)

    def Start(self):
        """
        start module
        """
        self.log.info("service starting")
        self.server.start()

    def Close(self):
        """
        close module
        """
        grace = None
        if 'timeout' in self.config['server']:
            grace = self.config['server']['timeout'] / 1e9
        self.server.stop(grace)
        self.log.info("service closed")

    def Call(self, request, context):
        """
        call request
        """

        if request.FunctionName not in self.functions:
            raise Exception('function not found')

        ctx = {}
        ctx['messageQOS'] = request.QOS
        ctx['messageTopic'] = request.Topic
        ctx['messageTimestamp'] = request.Timestamp
        ctx['functionName'] = request.FunctionName
        ctx['functionInvokeID'] = request.FunctionInvokeID
        ctx['invokeid'] = request.FunctionInvokeID

        msg = {}
        if request.Payload:
            try:
                msg = json.loads(request.Payload)
            except BaseException:
                msg = request.Payload  # raw data, not json format

        try:
            msg = self.functions[request.FunctionName](msg, ctx)
        except BaseException as err:
            self.log.error(err, exc_info=True)
            raise Exception("[UserCodeInvoke] ", err)

        if msg is None:
            request.Payload = b''
        elif isinstance(msg, bytes):
            request.Payload = msg
        else:
            try:
                request.Payload = json.dumps(msg).encode('utf-8')
            except BaseException as err:
                self.log.error(err, exc_info=True)
                raise Exception("[UserCodeReturn] ", err)
        return request

    def Talk(self, request_iterator, context):
        """
        talk request
        """
        pass


def get_functions(c):
    """
    get functions
    """
    fs = {}
    for fc in c:
        if 'name' not in fc or 'handler' not in fc or 'codedir' not in fc:
            raise Exception(
                'config invalid, missing function name, handler or codedir')
        sys.path.append(fc['codedir'])
        module_handler = fc['handler'].split('.')
        handler_name = module_handler.pop()
        module = importlib.import_module('.'.join(module_handler))
        fs[fc['name']] = getattr(module, handler_name)
    return fs


def get_grpc_server(c):
    """
    get grpc server
    """
    # TODO: to test
    max_workers = None
    max_concurrent = None
    max_message_length = 4 * 1024 * 1024
    if 'workers' in c:
        if 'max' in c['workers']:
            max_workers = c['workers']['max']
    if 'concurrent' in c:
        if 'max' in c['concurrent']:
            max_concurrent = c['concurrent']['max']
    if 'message' in c:
        if 'length' in c['message']:
            if 'max' in c['message']['length']:
                max_message_length = c['message']['length']['max']

    ssl_ca = None
    ssl_key = None
    ssl_cert = None
    if 'ca' in c:
        with open(c['ca'], 'rb') as f:
            ssl_ca = f.read()
    if 'key' in c:
        with open(c['key'], 'rb') as f:
            ssl_key = f.read()
    if 'cert' in c:
        with open(c['cert'], 'rb') as f:
            ssl_cert = f.read()

    s = grpc.server(thread_pool=futures.ThreadPoolExecutor(max_workers=max_workers),
                    options=[('grpc.max_send_message_length', max_message_length),
                             ('grpc.max_receive_message_length', max_message_length)],
                    maximum_concurrent_rpcs=max_concurrent)
    if ssl_key is not None and ssl_cert is not None:
        credentials = grpc.ssl_server_credentials(
            ((ssl_key, ssl_cert),), ssl_ca, ssl_ca is not None)
        s.add_secure_port(c['address'], credentials)
    else:
        s.add_insecure_port(c['address'])
    return s


def get_logger(c):
    """
    get logger
    """
    logger = logging.getLogger(c['name'])
    if 'logger' not in c:
        return logger

    if 'path' not in c['logger']:
        return logger

    filename = os.path.abspath(c['logger']['path'])
    os.makedirs(os.path.dirname(filename), exist_ok=True)

    level = logging.INFO
    if 'level' in c['logger']:
        if c['logger']['level'] == 'debug':
            level = logging.DEBUG
        elif c['logger']['level'] == 'warn':
            level = logging.WARNING
        elif c['logger']['level'] == 'error':
            level = logging.ERROR

    interval = 15
    if 'age' in c['logger'] and 'max' in c['logger']['age']:
        interval = c['logger']['age']['max']

    backupCount = 15
    if 'backup' in c['logger'] and 'max' in c['logger']['backup']:
        backupCount = c['logger']['backup']['max']

    logger.setLevel(level)

    # create a file handler
    handler = logging.handlers.TimedRotatingFileHandler(
        filename=filename,
        when='h',
        interval=interval,
        backupCount=backupCount)
    handler.setLevel(level)

    # create a logging format
    formatter = logging.Formatter(
        '%(asctime)s - %(name)s - %(levelname)s - %(message)s')
    handler.setFormatter(formatter)

    # add the handlers to the logger
    logger.addHandler(handler)
    return logger


if __name__ == '__main__':
    parser = argparse.ArgumentParser(
        description='grpc server for python3 function')
    parser.add_argument('-c',
                        type=str,
                        default=os.path.join("etc", "baetyl", "service.yml"),
                        help='config file path (default: etc/baetyl/service.yml)')
    args = parser.parse_args()
    m = mo()
    m.Load(args.c)
    m.Start()

    def exit(signum, frame):
        sys.exit(0)

    signal.signal(signal.SIGINT, exit)
    signal.signal(signal.SIGTERM, exit)

    try:
        while True:
            time.sleep(_ONE_DAY_IN_SECONDS)
    except BaseException as err:
        m.log.debug(err)
    finally:
        m.Close()
