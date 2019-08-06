#pragma once

#include <darknet.h>

extern detection * get_detection(detection *dets, int index, int dets_len);
extern float get_detection_probability(detection *det, int index, int prob_len);
