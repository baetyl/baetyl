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


if __name__ == "__main__":
    event = {"functionInvokeID": "8f2f5212-05ef-41d9-8e16-dd4974906f71",
             "imageCaptureTime": "2019-06-28T12:33:54.105352+08:00",
             "imageDiscard": False,
             "imageHight": 720,
             "imageInferenceTime": 0.232822577,
             "imageLocation": "1561696434105352000.jpg",
             "imageObjects": [{"class": "person",
                               "score": 0.6695935130119324,
                               "left": 0.1767835021018982,
                               "right": 0.8090300559997559,
                               "bottom": 0.9996803998947144,
                               "top": 0.2904333770275116}],
             "imageProcessTime": 0.27900074,
             "imageScores": {"person": 0.6695935130119324},
             "imageWidth": 1280,
             "messageTimestamp": 1561696434105,
             "publishTopic": "video/infer/result"}
    print(handler(event, {}))
