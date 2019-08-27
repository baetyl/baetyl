#!/usr/bin/env python
# -*- coding:utf-8 -*-
"""
function to mark image by video infer result in python
"""

import cv2


def handler(event, context):
    """
    function handler
    """
    if 'imageDiscard' in event and event['imageDiscard']:
        return None

    if 'imageLocation' not in event or len(event['imageLocation']) == 0:
        return None

    if 'imageObjects' not in event or len(event['imageObjects']) == 0:
        return None

    if 'imageWidth' not in event:
        return None

    if 'imageHight' not in event:
        return None

    img = cv2.imread(event['imageLocation'])
    font = cv2.FONT_HERSHEY_SIMPLEX  # FontHersheySimplex,FONT_HERSHEY_PLAIN
    text = ""
    if 'imageProcessTime' in event:
        text += " %.2ffps" % (1 / event['imageProcessTime'])
    if 'imageInferenceTime' in event:
        text += " infer %.0fms" % (event['imageInferenceTime'] * 1000)
    if text != "":
        cv2.putText(img, text, (0, 20), font, 0.5, (255, 0, 0), 1)

    objects = {}
    (w, h) = (event['imageWidth'], event['imageHight'])
    for obj in event['imageObjects']:
        (clazz, score) = (obj['class'], obj['score'])
        if score <= 0.6:
            continue
        objects[clazz] = score

        (left, top, right, bottom) = (
            int(obj['left'] * w), int(obj['top'] * h), int(obj['right'] * w), int(obj['bottom'] * h))
        cv2.rectangle(img, (left, top), (right, bottom), (0, 255, 0), 2)
        label = "%s %.2f" % (clazz, score * 100)
        text_size = cv2.getTextSize(label, font, 1, 2)
        text_width = text_size[0][0] + 2
        text_height = text_size[0][1] + 2
        if top < text_height:
            top = text_height
        cv2.rectangle(img, (left - 1, top - text_height),
                      (left - 1 + text_width, top), (0, 255, 0), -1)
        cv2.putText(img, label, (left, top), font, 1, (0, 0, 0), 2)

    loc = str(event['imageLocation'])
    loc_parts = loc.split(".")
    loc = "_".join(loc_parts[:-1])
    for k, v in objects.items():
        loc += '_' + k
    loc += ".jpg"
    cv2.imwrite(loc, img)

    return {'type': 'UPLOAD', 'content': {'localPath': loc, 'remotePath': loc}}
