FROM snowzach/doods:builder as build

ADD . .
RUN make

FROM ubuntu:18.04
RUN apt-get update && apt-get install -y --no-install-recommends libusb-1.0 libc++-7-dev wget unzip ca-certificates 
# For OpenCV
# RUN apt-get install -y --no-install-recommends libgtk2.0-0 libdc1394-22 libavcodec57 libavformat57 ffmpeg 
RUN mkdir -p /opt/doods
WORKDIR /opt/doods
COPY --from=build /usr/local/lib/. /usr/local/lib/.
COPY --from=build /build/doods /opt/doods/doods
ADD config.yaml /opt/doods/config.yaml
RUN ldconfig

RUN mkdir models
RUN wget https://storage.googleapis.com/download.tensorflow.org/models/tflite/coco_ssd_mobilenet_v1_1.0_quant_2018_06_29.zip && unzip coco_ssd_mobilenet_v1_1.0_quant_2018_06_29.zip && rm coco_ssd_mobilenet_v1_1.0_quant_2018_06_29.zip && mv detect.tflite models/coco_ssd_mobilenet_v1_1.0_quant.tflite && rm labelmap.txt
RUN wget https://dl.google.com/coral/canned_models/coco_labels.txt && mv coco_labels.txt models/coco_labels0.txt

CMD ["/opt/doods/doods", "-c", "/opt/doods/config.yaml", "api"]
