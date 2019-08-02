FROM ubuntu:18.04 as build

# Install reqs
RUN apt-get update && apt-get -y install pkg-config zip g++ zlib1g-dev unzip wget bash-completion git curl libusb-1.0 patch python python-future python3 libc++-7-dev

# Install Bazel
RUN wget https://github.com/bazelbuild/bazel/releases/download/0.27.0/bazel_0.27.0-linux-x86_64.deb && dpkg -i bazel_0.27.0-linux-x86_64.deb && rm bazel_0.27.0-linux-x86_64.deb

# Download and install the tensorflow lite shared library
RUN cd /opt && git clone https://github.com/tensorflow/tensorflow.git --branch r1.14 && \
    cd /opt/tensorflow && \
    bazel build -c opt --config monolithic --incompatible_no_support_tools_in_action_inputs=false tensorflow/lite:libtensorflowlite.so && \
    install bazel-out/k8-opt/bin/tensorflow/lite/libtensorflowlite.so /usr/local/lib/libtensorflowlite.so && \
    bazel build -c opt --config monolithic --incompatible_no_support_tools_in_action_inputs=false tensorflow/lite/experimental/c:libtensorflowlite_c.so && \
    install bazel-out/k8-opt/bin/tensorflow/lite/experimental/c/libtensorflowlite_c.so /usr/local/lib/libtensorflowlite_c.so && \
    mkdir -p /usr/local/include/flatbuffers && cp bazel-tensorflow/external/flatbuffers/include/flatbuffers/* /usr/local/include/flatbuffers && \
    bazel clean && rm -Rf /root/.cache

# Download the edgetpu library and install it
RUN cd /tmp && git clone https://coral.googlesource.com/edgetpu-native --branch release-chef && \
    install edgetpu-native/libedgetpu/libedgetpu_x86_64.so /usr/local/lib/libedgetpu.so && \
    mkdir -p /usr/local/include/libedgetpu && \
    install edgetpu-native/libedgetpu/edgetpu.h /usr/local/include/libedgetpu/edgetpu.h && \
    rm -Rf edgetpu-native
    
RUN ldconfig

RUN wget https://dl.google.com/go/go1.12.6.linux-amd64.tar.gz && tar -C /usr/local -xzf go1.12.6.linux-amd64.tar.gz && rm go1.12.6.linux-amd64.tar.gz
ENV PATH="/usr/local/go/bin:/go/bin:${PATH}"
ENV GOPATH="/go"

RUN wget https://github.com/protocolbuffers/protobuf/releases/download/v3.9.0/protoc-3.9.0-linux-x86_64.zip && unzip protoc-3.9.0-linux-x86_64.zip -d /usr/local && rm /usr/local/readme.txt && rm protoc-3.9.0-linux-x86_64.zip

# Create the build directory
RUN mkdir /build
WORKDIR /build
ADD . .

RUN make

FROM ubuntu:18.04
RUN apt-get update && apt-get -y install libusb-1.0 libc++-7-dev wget unzip
RUN mkdir -p /opt/doods
WORKDIR /opt/doods
COPY --from=build /usr/local/lib/libedgetpu.so /usr/local/lib/libedgetpu.so
COPY --from=build /usr/local/lib/libtensorflowlite.so /usr/local/lib/libtensorflowlite.so
COPY --from=build /usr/local/lib/libtensorflowlite_c.so /usr/local/lib/libtensorflowlite_c.so
COPY --from=build /build/doods /opt/doods/doods
RUN ldconfig

RUN wget https://storage.googleapis.com/download.tensorflow.org/models/tflite/coco_ssd_mobilenet_v1_1.0_quant_2018_06_29.zip && unzip coco_ssd_mobilenet_v1_1.0_quant_2018_06_29.zip && rm coco_ssd_mobilenet_v1_1.0_quant_2018_06_29.zip && mv detect.tflite model.tflite && rm labelmap.txt
RUN wget https://dl.google.com/coral/canned_models/mobilenet_ssd_v2_coco_quant_postprocess_edgetpu.tflite && mv mobilenet_ssd_v2_coco_quant_postprocess_edgetpu.tflite model-edgetpu.tflite
RUN wget https://dl.google.com/coral/canned_models/coco_labels.txt && mv coco_labels.txt labels.txt
CMD ["/opt/doods/doods","api"]
