package darknet

// #include <darknet.h>
//
// #include "detection.h"
import "C"
import "image"

// Detection represents a detection.
type Detection struct {
	BoundingBox

	ClassIDs      []int
	ClassNames    []string
	Probabilities []float32
}

func makeDetection(img *Image, det *C.detection, threshold float32, classes int,
	classNames []string) *Detection {
	dClassIDs := make([]int, 0)
	dClassNames := make([]string, 0)
	dProbs := make([]float32, 0)

	for i := 0; i < int(classes); i++ {
		dProb := float32(
			C.get_detection_probability(det, C.int(i), C.int(classes)))
		if dProb > threshold {
			dClassIDs = append(dClassIDs, i)
			cN := classNames[i]
			dClassNames = append(dClassNames, cN)
			dProbs = append(dProbs, dProb*100)
		}
	}

	fImgW := C.float(img.Width)
	fImgH := C.float(img.Height)
	halfRatioW := det.bbox.w / 2.0
	halfRatioH := det.bbox.h / 2.0
	out := Detection{
		BoundingBox: BoundingBox{
			StartPoint: image.Point{
				X: int((det.bbox.x - halfRatioW) * fImgW),
				Y: int((det.bbox.y - halfRatioH) * fImgH),
			},
			EndPoint: image.Point{
				X: int((det.bbox.x + halfRatioW) * fImgW),
				Y: int((det.bbox.y + halfRatioH) * fImgH),
			},
		},

		ClassIDs:      dClassIDs,
		ClassNames:    dClassNames,
		Probabilities: dProbs,
	}

	return &out
}

func makeDetections(img *Image, detections *C.detection, detectionsLength int,
	threshold float32, classes int, classNames []string) []*Detection {
	// Make list of detection objects.
	ds := make([]*Detection, detectionsLength)
	for i := 0; i < int(detectionsLength); i++ {
		det := C.get_detection(detections, C.int(i), C.int(classes))
		d := makeDetection(img, det, threshold, classes, classNames)
		ds[i] = d
	}

	return ds
}
