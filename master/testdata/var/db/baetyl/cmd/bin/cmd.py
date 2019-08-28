#!/usr/bin/env python
#

import time

if __name__ == '__main__':
    while True:
        try:
            time.sleep(10)
        except KeyboardInterrupt:
            break
