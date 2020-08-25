#!/bin/bash
# Check for edgetpu and specify devices if found
DOCKER_EXTRA=""
if `lsusb | egrep "(1a6e:089a|18d1:9302)" > /dev/null`; then
    DOCKER_EXTRA="--device /dev/bus/usb "
    echo "EdgeTPU detected..."
fi

docker run -it -v $PWD:/build -p 8090:8080 ${DOCKER_EXTRA} snowzach/doods:builder bash
