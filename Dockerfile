FROM ubuntu:18.04 as build

# Install reqs
RUN apt-get update && apt-get install -y --no-install-recommends \
    pkg-config zip g++ zlib1g-dev unzip wget bash-completion git curl \
    libusb-1.0 patch python python-future python3 libc++-7-dev \
    git build-essential cmake libgtk2.0-dev \
    ca-certificates libcurl4-openssl-dev libssl-dev \
    libavcodec-dev libavformat-dev libswscale-dev libtbb2 libtbb-dev \
    libjpeg-dev libpng-dev libtiff-dev libdc1394-22-dev && \
    rm -rf /var/lib/apt/lists/*

# Install Bazel
RUN wget https://github.com/bazelbuild/bazel/releases/download/0.27.0/bazel_0.27.0-linux-x86_64.deb && dpkg -i bazel_0.27.0-linux-x86_64.deb && rm bazel_0.27.0-linux-x86_64.deb

# Download and install the tensorflow lite shared library
RUN cd /opt && git clone https://github.com/tensorflow/tensorflow.git --branch r1.14 --single-branch && \
    cd /opt/tensorflow && \
    # tensorflow lite
    bazel build -c opt --config monolithic --incompatible_no_support_tools_in_action_inputs=false tensorflow/lite:libtensorflowlite.so && \
    install bazel-out/k8-opt/bin/tensorflow/lite/libtensorflowlite.so /usr/local/lib/libtensorflowlite.so && \
    bazel build -c opt --config monolithic --incompatible_no_support_tools_in_action_inputs=false tensorflow/lite/experimental/c:libtensorflowlite_c.so && \
    install bazel-out/k8-opt/bin/tensorflow/lite/experimental/c/libtensorflowlite_c.so /usr/local/lib/libtensorflowlite_c.so && \
    mkdir -p /usr/local/include/flatbuffers && cp bazel-tensorflow/external/flatbuffers/include/flatbuffers/* /usr/local/include/flatbuffers && \
    # tensorflow
    bazel build -c opt --config monolithic --incompatible_no_support_tools_in_action_inputs=false tensorflow:libtensorflow.so && \
    install bazel-out/k8-opt/bin/tensorflow/libtensorflow.so /usr/local/lib/libtensorflow.so && \
    ln -s /usr/local/lib/libtensorflow.so /usr/local/lib/libtensorflow.so.1 && \
    # cleanup
    bazel clean && rm -Rf /root/.cache

# Download the edgetpu library and install it
RUN cd /tmp && git clone https://coral.googlesource.com/edgetpu-native --branch release-chef && \
    install edgetpu-native/libedgetpu/libedgetpu_x86_64.so /usr/local/lib/libedgetpu.so && \
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
ARG GOVERSION="1.12.6"
ENV GOVERSION $GOVERSION
RUN apt-get update && apt-get install -y --no-install-recommends \
            git software-properties-common && \
            curl -Lo go${GOVERSION}.linux-amd64.tar.gz https://dl.google.com/go/go${GOVERSION}.linux-amd64.tar.gz && \
            tar -C /usr/local -xzf go${GOVERSION}.linux-amd64.tar.gz && \
            rm go${GOVERSION}.linux-amd64.tar.gz && \
            rm -rf /var/lib/apt/lists/*
ENV PATH /usr/local/go/bin:/go/bin:${PATH}
ENV GOPATH /go

# Install Protobuf compiler
RUN wget https://github.com/protocolbuffers/protobuf/releases/download/v3.9.0/protoc-3.9.0-linux-x86_64.zip && unzip protoc-3.9.0-linux-x86_64.zip -d /usr/local && rm /usr/local/readme.txt && rm protoc-3.9.0-linux-x86_64.zip

# Create the build directory
RUN mkdir /build
WORKDIR /build

# Copy everything above this for the builder image
ADD . .
RUN make

FROM ubuntu:18.04
RUN apt-get update && apt-get install -y --no-install-recommends libusb-1.0 libc++-7-dev wget unzip
RUN mkdir -p /opt/doods
WORKDIR /opt/doods
COPY --from=build /usr/local/lib/libedgetpu.so /usr/local/lib/libedgetpu.so
COPY --from=build /usr/local/lib/libtensorflowlite.so /usr/local/lib/libtensorflowlite.so
COPY --from=build /usr/local/lib/libtensorflowlite_c.so /usr/local/lib/libtensorflowlite_c.so
COPY --from=build /build/doods /opt/doods/doods
ADD config.toml /opt/doods/config.toml
RUN ldconfig

RUN wget https://storage.googleapis.com/download.tensorflow.org/models/tflite/coco_ssd_mobilenet_v1_1.0_quant_2018_06_29.zip && unzip coco_ssd_mobilenet_v1_1.0_quant_2018_06_29.zip && rm coco_ssd_mobilenet_v1_1.0_quant_2018_06_29.zip && mv detect.tflite models/coco_ssd_mobilenet_v1_1.0_quant.tflite && rm labelmap.txt
RUN wget https://dl.google.com/coral/canned_models/coco_labels.txt && mv coco_labels.txt models/coco_labels0.txt 

CMD ["/opt/doods/doods","-c", "/opt/doods/config.toml", "api"]
