ARG BUILDER_TAG="latest"
FROM registry.prozach.org/doods-builder:$BUILDER_TAG as builder

# Create the build directory
WORKDIR /build
ADD . .
RUN git status
RUN make
RUN git status

FROM debian:buster-slim
RUN apt-get update && \
    apt-get install -y --no-install-recommends libusb-1.0 libc++-7-dev wget unzip ca-certificates libdc1394-22 libavcodec58 libavformat58 && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*
RUN mkdir -p /opt/doods
WORKDIR /opt/doods
COPY --from=builder /usr/local/lib/. /usr/local/lib/.
COPY --from=builder /build/doods /opt/doods/doods
ADD config.yaml /opt/doods/config.yaml
RUN ldconfig

RUN mkdir models
RUN wget https://storage.googleapis.com/download.tensorflow.org/models/tflite/coco_ssd_mobilenet_v1_1.0_quant_2018_06_29.zip && unzip coco_ssd_mobilenet_v1_1.0_quant_2018_06_29.zip && rm coco_ssd_mobilenet_v1_1.0_quant_2018_06_29.zip && mv detect.tflite models/coco_ssd_mobilenet_v1_1.0_quant.tflite && rm labelmap.txt
RUN wget https://dl.google.com/coral/canned_models/coco_labels.txt && mv coco_labels.txt models/coco_labels0.txt

CMD ["/opt/doods/doods", "-c", "/opt/doods/config.yaml", "api"]
