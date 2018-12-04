#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
module to write data to TSDB
"""

import calendar
import datetime
import json
import requests
import time

from sign import sign


# set the transport protocol
TRANS_PROTOCOL = 'http://'

# set http method
HTTP_METHOD = 'POST'

# set base url and path
base_url = 'leenodb.tsdb.iot.gz.baidubce.com'
path = '/v1/datapoint'

# save the information of Access Key ID and Secret Access Key
ak = '7e97772725004a2baa5076c932f7f883'
sk = '90f3e1e530a944258774ca1ef51c33fa'
credentials = sign.BceCredentials(ak, sk)

# set a http header except field 'Authorization'
headers = {'Host': base_url,
           'Content-Type': 'application/json;charset=utf-8'}

# we don't have params in our url,so set it to None
params = None

# set header fields should be signed
headers_to_sign = {"host"}


def access_db(http_method, url, data=None):
    """
    function to access TSDB by RESTful API（only have GET,POST,PUT now)
    """

    # invoke sign method to get a signed string
    sign_str = sign.sign(credentials, HTTP_METHOD, path, headers, params,
                         headers_to_sign=headers_to_sign)

    # add field 'Authorization' to complete the whole http header
    final_headers = dict(headers.items() + {'Authorization': sign_str}.items())

    try:
        if (http_method == 'POST') and (data is not None):
            rsp = requests.post(url, headers=final_headers, data=json.dumps(data))
        elif http_method == 'GET':
            rsp = requests.get(url, headers=final_headers)
        elif (http_method == 'PUT') and (data is not None):
            rsp = requests.put(url, headers=final_headers, data=json.dumps(data))
        else:
            rsp = 'Bad http method or data is empty'
    except StandardError:
        raise

    return rsp


def build_datapoints(event):
    """
    function to build datapoints by event
    """

    datapoints = dict()

    datapoint = dict()
    datapoint['metric'] = 'temperature'
    datapoint['tags'] = {'unit': event['unit']}
    datapoint['value'] = event['temperature']
    timestamp = time.mktime(time.strptime(event['datetime'],
                            '%Y-%m-%d %H:%M:%S'))

    datapoint['timestamp'] = str(timestamp).split('.')[0]
    datapoints['datapoints'] = [datapoint]

    return datapoints


def handler(event, context):
    """
    function handler
    """

    datapoints = build_datapoints(event)

    try:
        rsp = access_db(HTTP_METHOD, TRANS_PROTOCOL + base_url + path,
                        datapoints)
    except StandardError:
        raise

    # check http response status code to confirm if we write data successfully
    if str(rsp.status_code) == '204':
        pass
    else:
        if isinstance(rsp, str):
            raise TypeError('Response must be a string')
        else:
            raise BaseException('Get error: ' + str(rsp.status_code))
