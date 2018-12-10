#!/usr/bin/env python2.7
#-*- coding:utf-8 -*-
"""
module to run function of python 2.7 as grpc server
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
import runtime_pb2
import runtime_pb2_grpc
import utils

_ONE_DAY_IN_SECONDS = 60 * 60 * 24


class mo(runtime_pb2_grpc.RuntimeServicer):
    """
    grpc server module for python2.7 function
    """

    def Load(self, conf):
        """
        load config and init module
        """
        conf = conf.strip()
        if conf[0] == "{" and conf[-1] == "}":
            self.config = json.loads(conf)
        else:
            self.config = yaml.load(open(conf, 'r').read())

        if 'name' not in self.config:
            raise Exception, 'Module config invalid, missing name'
        if 'server' not in self.config:
            raise Exception, 'Module config invalid, missing server'
        if 'function' not in self.config:
            raise Exception, 'Module config invalid, missing function'

        self.log = utils.get_logger(self.config)
        if 'name' not in self.config['function'] or 'handler' not in self.config['function'] or 'codedir' not in self.config['function']:
            raise Exception, 'Module config invalid, missing function name, handler or codedir'

        sys.path.append(self.config['function']['codedir'])
        module_handler = self.config['function']['handler'].split('.')
        handler_name = module_handler.pop()
        module = importlib.import_module('.'.join(module_handler))
        self.function = getattr(module, handler_name)

        max_message_size = 4 * 1024 * 1024
        if 'message' in self.config['server']:
            if 'length' in self.config['server']['message']:
                if 'max' in self.config['server']['message']['length']:
                    max_message_size = self.config['server']['message']['length']['max']
        self.server = grpc.server(thread_pool=futures.ThreadPoolExecutor(),
                                  options=[('grpc.max_send_message_length', max_message_size),
                                           ('grpc.max_receive_message_length', max_message_size)])
        runtime_pb2_grpc.add_RuntimeServicer_to_server(self, self.server)
        if 'address' not in self.config['server']:
            raise Exception, 'Module config invalid, missing server address'
        self.server.add_insecure_port(self.config['server']['address'])

    def Start(self):
        """
        start module
        """
        self.log.info("module starting")
        self.server.start()

    def Close(self):
        """
        close module
        """
        self.server.stop(self.config['server']['timeout'] / 1e9)
        self.log.info("module closed")

    def Handle(self, request, context):
        """
        handle request
        """
        ctx = {}
        ctx['messageQOS'] = request.QOS
        ctx['messageTopic'] = request.Topic
        ctx['functionName'] = request.FunctionName
        ctx['functionInvokeID'] = request.FunctionInvokeID
        ctx['invokeid'] = request.FunctionInvokeID
        if request.Payload:
            try:
                msg = json.loads(request.Payload)
            except ValueError:
                msg = request.Payload  # raw data, not json format
        msg = self.function(msg, ctx)
        if msg is None:
            request.Payload = b''
        else:
            request.Payload = json.dumps(msg)
        return request


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Python function server')
    parser.add_argument('-c',
                        type=str,
                        default=os.path.join("conf", "conf.yml"),
                        help='config file path (default: conf/conf.yml)')
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
    except BaseException as ex:
        m.log.debug(ex)
    finally:
        m.Close()
