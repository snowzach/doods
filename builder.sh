#!/bin/bash
docker run -it -v $PWD:/build --device /dev/bus/usb snowzach/doods:builder bash
