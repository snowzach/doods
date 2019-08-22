#!/bin/bash

# If anything fails, exit
set -e

if [ $# -lt 1 ]; then
    echo "$0 <arch> (k8,arm)"
    exit 1
fi


if [ "$1" == "latest" ] || [ "$1" == "x86avx" ] || [ "$1" == "gpu" ]; then

    if [ "$1" == "gpu" ]; then
        apt-get -y --no-install-recommends install gnupg # Needed for curl/apt-key to work
        echo "Installing NVidia Drivers..."
        curl -s -L https://nvidia.github.io/nvidia-docker/gpgkey | apt-key add -
        curl -s -L https://nvidia.github.io/nvidia-docker/ubuntu18.04/nvidia-docker.list | tee /etc/apt/sources.list.d/nvidia-docker.list
        apt-get update
        apt-get -y --no-install-recommends install nvidia-container-runtime 
    fi

    echo "Installing protoc 3.9.1 binaries..."
    wget https://github.com/protocolbuffers/protobuf/releases/download/v3.9.1/protoc-3.9.1-linux-x86_64.zip 
    unzip protoc-3.9.1-linux-x86_64.zip -d /usr/local
    rm /usr/local/readme.txt
    rm protoc-3.9.1-linux-x86_64.zip

    echo "Installing bazel 0.24.1 package..."
    wget https://github.com/bazelbuild/bazel/releases/download/0.24.1/bazel_0.24.1-linux-x86_64.deb
    dpkg -i bazel_0.24.1-linux-x86_64.deb
    rm bazel_0.24.1-linux-x86_64.deb

    exit 0
fi

if [ "$1" == "aarch64" ]; then
    echo "Installing protoc 3.9.1 binaries..."
    wget https://github.com/protocolbuffers/protobuf/releases/download/v3.9.1/protoc-3.9.1-linux-aarch_64.zip 
    unzip protoc-3.9.1-linux-aarch_64.zip -d /usr/local
    rm /usr/local/readme.txt
    rm protoc-3.9.1-linux-aarch_64.zip

    echo "Installing OpenJDK 8..."
    cd /tmp
    wget http://http.us.debian.org/debian/pool/main/o/openjdk-8/openjdk-8-jdk-headless_8u222-b10-1_arm64.deb
    wget http://http.us.debian.org/debian/pool/main/o/openjdk-8/openjdk-8-jre-headless_8u222-b10-1_arm64.deb
    dpkg -i --force-all openjdk-8-jdk-headless_8u222-b10-1_arm64.deb openjdk-8-jre-headless_8u222-b10-1_arm64.deb
    apt-get --fix-broken -y --no-install-recommends install
    rm openjdk-8-jdk-headless_8u222-b10-1_arm64.deb openjdk-8-jre-headless_8u222-b10-1_arm64.deb

    echo "Installing Bazel 0.24.1 Binary from https://github.com/PINTO0309/Bazel_bin..."
    cd /tmp
    wget --save-cookie /tmp/cookie "https://drive.google.com/uc?export=download&id=1SyfrRqX-6KF_KxD4PBVQLT35tgI8Sf58" > /dev/null
    CODE="$(awk '/_warning_/ {print $NF}' /tmp/cookie)"
    wget --load-cookie /tmp/cookie -O bazel "https://drive.google.com/uc?export=download&confirm=${CODE}&id=1SyfrRqX-6KF_KxD4PBVQLT35tgI8Sf58"
    install bazel /usr/bin/bazel
    rm bazel

    exit 0
fi

if [ "$1" == "pi" ]; then
    echo "Compiling protoc 3.9.1 from source..."
    cd /tmp
    wget https://github.com/protocolbuffers/protobuf/releases/download/v3.9.1/protobuf-java-3.9.1.zip
    unzip protobuf-java-3.9.1.zip
    cd /tmp/protobuf-3.9.1
    ./configure
    make
    make install
    ldconfig
    rm -Rf /tmp/protobuf-3.9.1

    echo "Installing OpenJDK 8..."
    cd /tmp
    wget http://http.us.debian.org/debian/pool/main/o/openjdk-8/openjdk-8-jdk-headless_8u222-b10-1_armhf.deb
    wget http://http.us.debian.org/debian/pool/main/o/openjdk-8/openjdk-8-jre-headless_8u222-b10-1_armhf.deb
    dpkg -i --force-all openjdk-8-jdk-headless_8u222-b10-1_armhf.deb openjdk-8-jre-headless_8u222-b10-1_armhf.deb
    apt-get --fix-broken -y --no-install-recommends install
    rm openjdk-8-jdk-headless_8u222-b10-1_armhf.deb openjdk-8-jre-headless_8u222-b10-1_armhf.deb

    echo "Installing Bazel 0.24.1 Binary from https://github.com/PINTO0309/Bazel_bin..."
    cd /tmp
    wget --save-cookie /tmp/cookie "https://drive.google.com/uc?export=download&id=1iTm4fmCOxKDGTf6k3J_ubAyYdoV1I7LL" > /dev/null
    CODE="$(awk '/_warning_/ {print $NF}' /tmp/cookie)"
    wget --load-cookie /tmp/cookie -O bazel "https://drive.google.com/uc?export=download&confirm=${CODE}&id=1iTm4fmCOxKDGTf6k3J_ubAyYdoV1I7LL"
    install bazel /usr/bin/bazel
    rm bazel
    exit 0

fi

echo "Unknown arch $1"
exit 1
