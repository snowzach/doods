#include <stdlib.h>

#include <darknet.h>

#include "network.h"

int get_network_layer_classes(network *n, int index)
{
	return n->layers[index].classes;
}

struct network_box_result perform_network_detect(network *n, image *img,
    int classes, float thresh, float hier_thresh, float nms)
{
    image sized = letterbox_image(*img, n->w, n->h);

    struct network_box_result result = { NULL };

    float *X = sized.data;
    network_predict(n, X);

    result.detections = get_network_boxes(n, img->w, img->h,
        thresh, hier_thresh, 0, 1, &result.detections_len);
    if (nms) {
        do_nms_sort(result.detections, result.detections_len, classes, nms);
    }

    free_image(sized);

    return result;
}
