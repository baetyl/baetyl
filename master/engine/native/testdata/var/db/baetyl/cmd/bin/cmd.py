#!/usr/bin/env python
#

import time
import signal

if __name__ == '__main__':
    def exit(signum, frame):
        sys.exit(0)
    signal.signal(signal.SIGINT, exit)
    signal.signal(signal.SIGTERM, exit)

    while True:
         time.sleep(10)