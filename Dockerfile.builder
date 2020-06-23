FROM ubuntu:18.04 as base

# Install reqs with cross compile support
RUN apt-get update && apt-get install -y --no-install-recommends \
    pkg-config zip zlib1g-dev unzip wget bash-completion git curl \
    build-essential patch g++ python python-future python-numpy python-six python3 \
    cmake ca-certificates \
    libc6-dev libstdc++6 libusb-1.0-0

# Install protoc
RUN wget https://github.com/protocolbuffers/protobuf/releases/download/v3.12.3/protoc-3.12.3-linux-x86_64.zip && \
    unzip protoc-3.12.3-linux-x86_64.zip -d /usr/local && \
    rm /usr/local/readme.txt && \
    rm protoc-3.12.3-linux-x86_64.zip

# Version Configuration
ARG BAZEL_VERSION="2.0.0"
ARG TF_VERSION="f394a768719a55b5c351ed1ecab2ec6f16f99dd4"
ARG OPENCV_VERSION="4.3.0"
ARG GO_VERSION="1.14.3"

# Install bazel
ENV BAZEL_VERSION $BAZEL_VERSION
RUN wget https://github.com/bazelbuild/bazel/releases/download/${BAZEL_VERSION}/bazel_${BAZEL_VERSION}-linux-x86_64.deb && \
    dpkg -i bazel_${BAZEL_VERSION}-linux-x86_64.deb && \
    rm bazel_${BAZEL_VERSION}-linux-x86_64.deb

# Download tensorflow sources
ENV TF_VERSION $TF_VERSION
#RUN cd /opt && git clone https://github.com/tensorflow/tensorflow.git --branch $TF_VERSION --single-branch
RUN cd /opt && git clone https://github.com/tensorflow/tensorflow.git && cd /opt/tensorflow && git checkout ${TF_VERSION}

# Configure tensorflow
ENV TF_NEED_GDR=0 TF_NEED_AWS=0 TF_NEED_GCP=0 TF_NEED_CUDA=0 TF_NEED_HDFS=0 TF_NEED_OPENCL_SYCL=0 TF_NEED_VERBS=0 TF_NEED_MPI=0 TF_NEED_MKL=0 TF_NEED_JEMALLOC=1 TF_ENABLE_XLA=0 TF_NEED_S3=0 TF_NEED_KAFKA=0 TF_NEED_IGNITE=0 TF_NEED_ROCM=0
RUN cd /opt/tensorflow && yes '' | ./configure

# Tensorflow build flags
ENV BAZEL_COPT_FLAGS="--local_resources 16000,16,1 -c opt --config monolithic --copt=-march=native --copt=-O3 --copt=-fomit-frame-pointer --incompatible_no_support_tools_in_action_inputs=false --config=noaws --config=nohdfs"
ENV BAZEL_EXTRA_FLAGS="--host_linkopt=-lm"

# Compile and build tensorflow lite
RUN cd /opt/tensorflow && \
    bazel build -c opt $BAZEL_COPT_FLAGS --verbose_failures $BAZEL_EXTRA_FLAGS //tensorflow/lite:libtensorflowlite.so && \
    install bazel-bin/tensorflow/lite/libtensorflowlite.so /usr/local/lib/libtensorflowlite.so && \
    bazel build -c opt $BAZEL_COPT_FLAGS --verbose_failures $BAZEL_EXTRA_FLAGS //tensorflow/lite/c:libtensorflowlite_c.so && \
    install bazel-bin/tensorflow/lite/c/libtensorflowlite_c.so /usr/local/lib/libtensorflowlite_c.so && \
    mkdir -p /usr/local/include/flatbuffers && cp bazel-tensorflow/external/flatbuffers/include/flatbuffers/* /usr/local/include/flatbuffers

# Compile and install tensorflow shared library
RUN cd /opt/tensorflow && \
    bazel build -c opt $BAZEL_COPT_FLAGS --verbose_failures $BAZEL_EXTRA_FLAGS //tensorflow:libtensorflow.so && \
    install bazel-bin/tensorflow/libtensorflow.so /usr/local/lib/libtensorflow.so && \
    ln -rs /usr/local/lib/libtensorflow.so /usr/local/lib/libtensorflow.so.1 && \
    ln -rs /usr/local/lib/libtensorflow.so /usr/local/lib/libtensorflow.so.2

# cleanup so the cache directory isn't huge
RUN cd /opt/tensorflow && \
    bazel clean && rm -Rf /root/.cache

# Install GOCV
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
    make preinstall && make install && \
    cd /tmp && rm -rf opencv*

# Fetch the edgetpu library locally
ADD libedgetpu/out/throttled/k8/libedgetpu.so.1.0 /usr/local/lib/libedgetpu.so.1.0
RUN ln -rs /usr/local/lib/libedgetpu.so.1.0 /usr/local/lib/libedgetpu.so.1 && \
    ln -rs /usr/local/lib/libedgetpu.so.1.0 /usr/local/lib/libedgetpu.so && \
    mkdir -p /usr/local/include/libedgetpu
ADD libedgetpu/tflite/public/edgetpu.h /usr/local/include/libedgetpu/edgetpu.h
ADD libedgetpu/tflite/public/edgetpu_c.h /usr/local/include/libedgetpu/edgetpu_c.h

# Configure the Go version to be used
ENV GO_ARCH "amd64"
ENV GOARCH=amd64

# Install Go
ENV GO_VERSION $GO_VERSION
RUN curl -kLo go${GO_VERSION}.linux-${GO_ARCH}.tar.gz https://dl.google.com/go/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz && \
    tar -C /usr/local -xzf go${GO_VERSION}.linux-${GO_ARCH}.tar.gz && \
    rm go${GO_VERSION}.linux-${GO_ARCH}.tar.gz

FROM ubuntu:18.04 as builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    pkg-config zip zlib1g-dev unzip wget bash-completion git curl \
    build-essential patch g++ python python-future python3 ca-certificates \
    libc6-dev libstdc++6 libusb-1.0-0

# Copy all libraries, includes and go
COPY --from=base /usr/local/. /usr/local/.
COPY --from=base /opt/tensorflow /opt/tensorflow 

ENV GOOS=linux
ENV CGO_ENABLED=1
ENV CGO_CFLAGS=-I/opt/tensorflow
ENV PATH /usr/local/go/bin:/go/bin:${PATH}
ENV GOPATH /go

# Switch to /build
WORKDIR /build
