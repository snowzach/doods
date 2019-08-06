#pragma once

#include <darknet.h>

struct network_box_result {
    detection *detections;
    int detections_len;
};

extern int get_network_layer_classes(network *n, int index);
extern struct network_box_result perform_network_detect(
    network *n, image *img,
    int classes, float thresh, float hier_thresh, float nms);
