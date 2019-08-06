#include <darknet.h>

detection * get_detection(detection *dets, int index, int dets_len)
{
    if (index >= dets_len) {
        return NULL;
    }

    return dets + index;
}

float get_detection_probability(detection *det, int index, int prob_len)
{
    if (index >= prob_len) {
        return .0;
    }

    return det->prob[index];
}
