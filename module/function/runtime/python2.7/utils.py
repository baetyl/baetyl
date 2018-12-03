"""
utils
"""

import logging
import logging.handlers

def get_logger(c):
    """
    get logger
    """
    logger = logging.getLogger(c['name'])
    if 'logger' not in c:
        return logger

    if 'path' not in c['logger']:
        c['logger']['path'] = "var/log/" + c['name'] + ".log"

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
        c['logger']['path'], when='h', interval=interval, backupCount=backupCount)
    handler.setLevel(level)

    # create a logging format
    formatter = logging.Formatter(
        '%(asctime)s - %(name)s - %(levelname)s - %(message)s')
    handler.setFormatter(formatter)

    # add the handlers to the logger
    logger.addHandler(handler)
    return logger
