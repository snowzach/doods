#!/bin/bash

if [ $# -lt 1 ]; then
    "$0 builder/configs/<config>.conf"
    exit 1
fi

# Make sure the config exists
if [ ! -f "builder/configs/$1.conf" ]; then
    echo "Could not find builder/configs/$1.conf"
    exit 1
fi

# Read the build args from file
oldifs="$IFS"
IFS=$'\n\r'
BUILDARGS=""
for arg in `grep -v '^#' builder/configs/$1.conf`; do
    BUILDARGS+="--build-arg ${arg} "
done
IFS="$oldifs"

# need eval to properly parse up arguments
eval docker build ${BUILDARGS} --build-arg TAG=$1 -t snowzach/doods-builder:$1 -f builder/Dockerfile .
