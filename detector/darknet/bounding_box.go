package darknet

import (
	"image"
)

// BoundingBox represents a bounding box.
type BoundingBox struct {
	StartPoint image.Point
	EndPoint   image.Point
}
