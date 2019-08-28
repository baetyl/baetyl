#!/usr/bin/env python
# -*- coding:utf-8 -*-
"""
function to analyse video infer result in python
"""

import numpy as np

location = "var/db/baetyl/image/{}.jpg"
classes = {
    1: 'person'
}


def handler(event, context):
    """
    function handler
    """

    # mat = np.array(event, np.float32)
    data = np.fromstring(event, np.float32)
    mat = np.reshape(data, (-1, 7))
    objects = []
    scores = {}
    for obj in mat:
        clazz = int(obj[1])
        if clazz in classes:
            score = float(obj[2])
            if classes[clazz] not in scores or scores[classes[clazz]] < score:
                scores[classes[clazz]] = score
            if score < 0.6:
                continue
            objects.append({
                'class': classes[clazz],
                'score': score,
                'left': float(obj[3]),
                'top': float(obj[4]),
                'right': float(obj[5]),
                'bottom': float(obj[6])
            })

    res = {}
    res["imageDiscard"] = len(objects) == 0
    res["imageObjects"] = objects
    res["imageScores"] = scores
    res["publishTopic"] = "video/infer/result"
    res["messageTimestamp"] = int(context["messageTimestamp"]/1000000)
    if len(objects) != 0:
        res["imageLocation"] = location.format(context["messageTimestamp"])

    return res