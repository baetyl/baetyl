#!/usr/bin/env python
# -*- coding:utf-8 -*-
"""
module to say hi
"""


def handler(event, context):
    """
    function handler
    """ 
    res = {}
    if isinstance(event, dict):
        if "err" in event:
            raise TypeError(event['err'])
        res = event
    elif isinstance(event, bytes):
        res['bytes'] = event.decode("utf-8")

    if 'messageQOS' in context:
        res['messageQOS'] = context['messageQOS']
    if 'messageTopic' in context:
        res['messageTopic'] = context['messageTopic']
    if 'functionName' in context:
        res['functionName'] = context['functionName']
    if 'functionInvokeID' in context:
        res['functionInvokeID'] = context['functionInvokeID']

    res['Say'] = 'Hello OpenEdge'
    return res
