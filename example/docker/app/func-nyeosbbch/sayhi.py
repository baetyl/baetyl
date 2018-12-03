#!/usr/bin/env python
#-*- coding:utf-8 -*-
"""
module to say hi
"""

import os
import time
import threading


def handler(event, context):
    """
    function handler
    """
    if 'i' in event:
        if event['i'] > 10:
            return None

    if 't' in event:
        time.sleep(event['t'])
        event['sleep'] = True
        return event

    if 'e' in event:
        event['e'] = 1 / 0

    if 's' in event:
        size = event['s']  # MB
        data = ' ' * (size * 1024 * 1024)
        event['l'] = len(data)

    if 'f' in event:
        try:
            f_o = open('sayhi.txt', 'w')
            f_o.write('Hello World')
            event['f_w_n'] = f_o.name
            f_o.close()
        except BaseException as ex:
            event['f_w_e'] = str(ex)
        try:
            f_o = open('../../conf/openedge.yml', 'r')
            event['f_r_n'] = f_o.name
            event['f_r_d'] = f_o.read()[:10]
            f_o.close()
        except BaseException as ex:
            event['f_r_e'] = str(ex)

    if 'p' in event:
        thr = threading.Thread(target=run)
        thr.setDaemon(True)
        thr.start()
        time.sleep(5)

    if 'c' in event:
        while True:
            pass

    if 'invoke' in event:
        res = context.invoke(event['invoke'], event['invokeArgs'])
        res['invoked'] = True
        return res

    event['USER_ID'] = os.environ['USER_ID']
    event['functionName'] = context['functionName']
    event['functionInvokeID'] = context['functionInvokeID']
    event['invokeid'] = context['invokeid']
    event['messageQOS'] = context['messageQOS']
    event['messageTopic'] = context['messageTopic']
    event['py'] = '你好，世界！'

    return event


def run(event):
    """
    function run thread
    """
    for i in range(1, 10):
        event['run.thread.times'] = i
        time.sleep(5)
