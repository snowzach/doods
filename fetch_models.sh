#!/bin/bash
mkdir -p models
# coco_ssd_mobilenet_v1_1.0_quant_2018_06_29
wget https://storage.googleapis.com/download.tensorflow.org/models/tflite/coco_ssd_mobilenet_v1_1.0_quant_2018_06_29.zip && unzip coco_ssd_mobilenet_v1_1.0_quant_2018_06_29.zip && rm coco_ssd_mobilenet_v1_1.0_quant_2018_06_29.zip && mv detect.tflite models/coco_ssd_mobilenet_v1_1.0_quant.tflite && rm labelmap.txt
wget https://dl.google.com/coral/canned_models/coco_labels.txt && mv coco_labels.txt models/coco_labels0.txt 
# mobilenet_ssd_v2_coco_quant_postprocess_edgetpu
wget https://dl.google.com/coral/canned_models/mobilenet_ssd_v2_coco_quant_postprocess_edgetpu.tflite && mv mobilenet_ssd_v2_coco_quant_postprocess_edgetpu.tflite models/mobilenet_ssd_v2_coco_quant_postprocess_edgetpu.tflite
# faster_rcnn_inception_v2_coco_2018_01_28
wget http://download.tensorflow.org/models/object_detection/faster_rcnn_inception_v2_coco_2018_01_28.tar.gz && tar -zxvf faster_rcnn_inception_v2_coco_2018_01_28.tar.gz faster_rcnn_inception_v2_coco_2018_01_28/frozen_inference_graph.pb --strip=1 && mv frozen_inference_graph.pb models/faster_rcnn_inception_v2_coco_2018_01_28.pb && rm faster_rcnn_inception_v2_coco_2018_01_28.tar.gz
wget https://raw.githubusercontent.com/amikelive/coco-labels/master/coco-labels-2014_2017.txt && mv coco-labels-2014_2017.txt models/coco_labels1.txt

cat << EOF > example.yaml
doods:
  detectors:
    - name: default
      type: tflite
      modelFile: models/coco_ssd_mobilenet_v1_1.0_quant.tflite
      labelFile: models/coco_labels0.txt
      numThreads: 0
      numConcurrent: 4 
    - name: edgetpu
      type: tflite
      modelFile: models/mobilenet_ssd_v2_coco_quant_postprocess_edgetpu.tflite
      labelFile: models/coco_labels0.txt
      numThreads: 0
      numConcurrent: 4
      hwAccel: true
    - name: tensorflow
      type: tensorflow
      modelFile: models/faster_rcnn_inception_v2_coco_2018_01_28.pb
      labelFile: models/coco_labels1.txt
      width: 224
      height: 224
      numThreads: 0
      numConcurrent: 4 
EOF