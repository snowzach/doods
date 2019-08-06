package darknet

import "time"

// DetectionResult represents the inference results from the network.
type DetectionResult struct {
	Detections           []*Detection
	NetworkOnlyTimeTaken time.Duration
	OverallTimeTaken     time.Duration
}
