package detector

import (
	"fmt"

	"github.com/snowzach/doods/odrpc"
)

func (m *Mux) FilterResponse(request *odrpc.DetectRequest, response *odrpc.DetectResponse) {

	// No filters, return everything
	if len(request.Detect) == 0 && len(request.Regions) == 0 {
		return
	}

	temp := response.Detections[:0]

detectionsLoop:
	for _, detection := range response.Detections {
		// Cleanup the bounds
		if detection.Top < 0 {
			detection.Top = 0
		}
		if detection.Left < 0 {
			detection.Left = 0
		}
		if detection.Bottom > 1 {
			detection.Bottom = 1
		}
		if detection.Right > 1 {
			detection.Right = 1
		}

		// We have this class listed explicitly
		if score, ok := request.Detect[detection.Label]; ok {
			if detection.Confidence >= score {
				temp = append(temp, detection)
				continue
			}
			// Wildcard class
		} else if score, ok := request.Detect["*"]; ok {
			if detection.Confidence >= score {
				temp = append(temp, detection)
				continue
			}
		}

		for _, region := range request.Regions {
			var inRegion bool
			if region.Covers {
				if detection.Top >= region.Top && detection.Left >= region.Left && detection.Bottom <= region.Bottom && detection.Right <= region.Right {
					inRegion = true
				}
			} else {
				if detection.Top <= region.Bottom && detection.Left <= region.Right && detection.Bottom >= region.Top && detection.Right >= region.Left {
					inRegion = true
				}
			}
			if inRegion {
				// We have this class listed explicitly
				if score, ok := region.Detect[detection.Label]; ok {
					if detection.Confidence >= score {
						temp = append(temp, detection)
						continue detectionsLoop
					}
					// Wildcard class
				} else if score, ok := region.Detect["*"]; ok {
					if detection.Confidence >= score {
						temp = append(temp, detection)
						continue detectionsLoop
					}
				}
			}
		}
	}

	response.Detections = temp
	for _, detection := range response.Detections {
		m.logger.Debugw("Detection", "id", request.Id, "label", detection.Label, "confidence", detection.Confidence, "location", fmt.Sprintf("%f,%f,%f,%f", detection.Top, detection.Left, detection.Bottom, detection.Right))
	}
}
