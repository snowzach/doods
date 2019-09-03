FROM debian:buster as builder

# The build tag
ARG TAG="latest"
ENV TAG $TAG
ARG DEPS_ARCH="x86_64"

# Install reqs
RUN apt-get update && apt-get install -y --no-install-recommends \
    pkg-config zip g++ zlib1g-dev unzip wget bash-completion git curl \
    libusb-1.0 patch python python-future python3 libc++-7-dev \
    git build-essential cmake ca-certificates libcurl4-openssl-dev libssl-dev \
    libavcodec-dev libavformat-dev libswscale-dev libtbb2 libtbb-dev \
    libjpeg-dev libpng-dev libtiff-dev libdc1394-22-dev swig software-properties-common

ARG TF_VERSION="r1.14"
ENV TF_VERSION $TF_VERSION
ARG TF_BAZEL_ARCH="k8"
ARG TF_BAZEL_OPTS=""
ARG TF_BAZEL_BUILD_OPTS="-c opt --config monolithic --incompatible_no_support_tools_in_action_inputs=false"

# Install Dependancies (Bazel, Protobuf, GPU Support, etc based on TAG)
ADD builder/install_deps.sh /tmp/install_deps.sh
RUN /tmp/install_deps.sh $DEPS_ARCH $TAG && rm /tmp/install_deps.sh

# Download tensorflow sources
RUN cd /opt && git clone https://github.com/tensorflow/tensorflow.git --branch $TF_VERSION --single-branch
# Compile and install tensorflow lite
RUN cd /opt/tensorflow && \
    bazel $TF_BAZEL_OPTS build $TF_BAZEL_BUILD_OPTS tensorflow/lite:libtensorflowlite.so && \
    install bazel-out/$TF_BAZEL_ARCH-opt/bin/tensorflow/lite/libtensorflowlite.so /usr/local/lib/libtensorflowlite.so && \
    bazel $TF_BAZEL_OPTS build $TF_BAZEL_BUILD_OPTS tensorflow/lite/experimental/c:libtensorflowlite_c.so && \
    install bazel-out/$TF_BAZEL_ARCH-opt/bin/tensorflow/lite/experimental/c/libtensorflowlite_c.so /usr/local/lib/libtensorflowlite_c.so && \
    mkdir -p /usr/local/include/flatbuffers && cp bazel-tensorflow/external/flatbuffers/include/flatbuffers/* /usr/local/include/flatbuffers
# Compile and install tensorflow    
RUN cd /opt/tensorflow && \
    bazel $TF_BAZEL_OPTS build $TF_BAZEL_BUILD_OPTS tensorflow:libtensorflow.so && \
    install bazel-out/$TF_BAZEL_ARCH-opt/bin/tensorflow/libtensorflow.so /usr/local/lib/libtensorflow.so && \
    ln -s /usr/local/lib/libtensorflow.so /usr/local/lib/libtensorflow.so.1
    # cleanup
RUN cd /opt/tensorflow && \
    bazel clean && rm -Rf /root/.cache

ARG EDGETPU_ARCH="x86_64"

# Download the edgetpu library and install it
RUN cd /tmp && git clone https://coral.googlesource.com/edgetpu-native --branch release-chef && \
    install edgetpu-native/libedgetpu/libedgetpu_$EDGETPU_ARCH.so /usr/local/lib/libedgetpu.so && \
    mkdir -p /usr/local/include/libedgetpu && \
    install edgetpu-native/libedgetpu/edgetpu.h /usr/local/include/libedgetpu/edgetpu.h && \
    rm -Rf edgetpu-native

# Install GOCV
ARG OPENCV_VERSION="4.0.1"
ENV OPENCV_VERSION $OPENCV_VERSION
RUN cd /tmp && \
    curl -Lo opencv.zip https://github.com/opencv/opencv/archive/${OPENCV_VERSION}.zip && \
    unzip -q opencv.zip && \
    curl -Lo opencv_contrib.zip https://github.com/opencv/opencv_contrib/archive/${OPENCV_VERSION}.zip && \
    unzip -q opencv_contrib.zip && \
    rm opencv.zip opencv_contrib.zip && \
    cd opencv-${OPENCV_VERSION} && \
    mkdir build && cd build && \
    cmake -D CMAKE_BUILD_TYPE=RELEASE \
            -D CMAKE_INSTALL_PREFIX=/usr/local \
            -D OPENCV_EXTRA_MODULES_PATH=../../opencv_contrib-${OPENCV_VERSION}/modules \
            -D WITH_JASPER=OFF \
            -D WITH_QT=OFF \
            -D WITH_GTK=OFF \
            -D BUILD_DOCS=OFF \
            -D BUILD_EXAMPLES=OFF \
            -D BUILD_TESTS=OFF \
            -D BUILD_PERF_TESTS=OFF \
            -D BUILD_opencv_java=NO \
            -D BUILD_opencv_python=NO \
            -D BUILD_opencv_python2=NO \
            -D BUILD_opencv_python3=NO \
            -D OPENCV_GENERATE_PKGCONFIG=ON .. && \
    make -j $(nproc --all) && \
    make preinstall && make install && ldconfig && \
    cd /tmp && rm -rf opencv*

# Install Go
ARG GO_ARCH="amd64"
ARG GO_VERSION="1.12.9"
ENV GO_VERSION $GO_VERSION
RUN curl -Lo go${GO_VERSION}.linux-$GO_ARCH.tar.gz https://dl.google.com/go/go${GO_VERSION}.linux-$GO_ARCH.tar.gz && \
    tar -C /usr/local -xzf go${GO_VERSION}.linux-$GO_ARCH.tar.gz && \
    rm go${GO_VERSION}.linux-$GO_ARCH.tar.gz
ENV PATH /usr/local/go/bin:/go/bin:${PATH}
ENV GOPATH /go

# Make the build container slightly smaller
RUN apt-get clean && rm -rf /var/lib/apt/lists/*
