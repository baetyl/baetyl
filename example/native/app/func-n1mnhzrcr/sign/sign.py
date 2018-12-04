#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
module to sign headers
"""

import hashlib
import hmac
import string
import datetime
import calendar


AUTHORIZATION = "authorization"
BCE_PREFIX = "x-bce-"
DEFAULT_ENCODING = 'UTF-8'


class BceCredentials(object):
    """
    class to save access_key_id and secret_access_key
    """

    def __init__(self, access_key_id, secret_access_key):
        self.access_key_id = access_key_id
        self.secret_access_key = secret_access_key


RESERVED_CHAR_SET = set(string.ascii_letters + string.digits + '.~-_')


def __get_normalized_char(i):
    char = chr(i)
    if char in RESERVED_CHAR_SET:
        return char
    else:
        return '%%%02X' % i


NORMALIZED_CHAR_LIST = [__get_normalized_char(i) for i in range(256)]



def __normalize_string(in_str, encoding_slash=True):
    """
    formalize string
    """

    if in_str is None:
        return ''

    in_str = in_str.encode(DEFAULT_ENCODING) if isinstance(in_str, unicode) else str(in_str)

    if encoding_slash:
        encode_f = lambda c: NORMALIZED_CHAR_LIST[ord(c)]
    else:
        encode_f = lambda c: NORMALIZED_CHAR_LIST[ord(c)] if c != '/' else c

    return ''.join([encode_f(ch) for ch in in_str])


def __get_canonical_time(timestamp=0):
    """
    get formalization timeStamp
    """

    if timestamp == 0:
        utctime = datetime.datetime.utcnow()
    else:
        utctime = datetime.datetime.utcfromtimestamp(timestamp)

    return "%04d-%02d-%02dT%02d:%02d:%02dZ" % (
        utctime.year, utctime.month, utctime.day,
        utctime.hour, utctime.minute, utctime.second)


def __get_canonical_uri(path):
    """
    get formalization URI
    """

    return __normalize_string(path, False)


def __get_canonical_querystring(params):
    """
    get formalization query string
    """

    if params is None:
        return ''

    result = ['%s=%s' % (k, __normalize_string(v))
              for k, v in params.items()
              if k.lower != AUTHORIZATION]

    result.sort()

    return '&'.join(result)


def __get_canonical_headers(headers, headers_to_sign=None):
    """
    get formalization header
    """

    headers = headers or {}

    if headers_to_sign is None or len(headers_to_sign) == 0:
        headers_to_sign = {"host", "content-md5", "content-length", "content-type"}

    f = lambda (key, value): (key.strip().lower(), str(value).strip())

    result = []
    for k, v in map(f, headers.iteritems()):
        if k.startswith(BCE_PREFIX) or k in headers_to_sign:
            result.append("%s:%s" % (__normalize_string(k), __normalize_string(v)))

    result.sort()

    return '\n'.join(result)


def sign(credentials, http_method, path, headers, params,
         timestamp=0, expiration_in_seconds=1800, headers_to_sign=None):
    """
    method to sign some filed in header
    """
    headers = headers or {}
    params = params or {}

    sign_key_info = 'bce-auth-v1/%s/%s/%d' % (
        credentials.access_key_id,
        __get_canonical_time(timestamp),
        expiration_in_seconds)

    sign_key = hmac.new(
        str(credentials.secret_access_key),
        str(sign_key_info),
        hashlib.sha256).hexdigest()

    canonical_uri = __get_canonical_uri(path)

    canonical_querystring = __get_canonical_querystring(params)

    canonical_headers = __get_canonical_headers(headers, headers_to_sign)

    string_to_sign = '\n'.join(
        [http_method, canonical_uri, canonical_querystring, canonical_headers])

    sign_result = hmac.new(sign_key, string_to_sign, hashlib.sha256).hexdigest()

    if headers_to_sign:
        result = '%s/%s/%s' % (sign_key_info, ';'.join(headers_to_sign), sign_result)
    else:
        result = '%s//%s' % (sign_key_info, sign_result)

    return result
