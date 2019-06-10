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

    if 'functionName' in context:
        event['functionName'] = context['functionName']
    if 'functionInvokeID' in context:
        event['functionInvokeID'] = context['functionInvokeID']
    if 'messageQOS' in context:
        event['messageQOS'] = context['messageQOS']
    if 'messageTopic' in context:
        event['messageTopic'] = context['messageTopic']
    event['py'] = 'hello world'

    return event


def run(event):
    """
    function run thread
    """
    for i in range(1, 10):
        event['run.thread.times'] = i
        time.sleep(5)
