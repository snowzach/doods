package darknet

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"github.com/snowzach/doods/detector/dconfig"
	"github.com/snowzach/doods/odrpc"
)

const YOLO3_FASTEST = 320

type detector struct {
	config odrpc.Detector
	logger *zap.SugaredLogger

	labels map[int]string
	pool   chan *YOLONetwork
}

func New(c *dconfig.DetectorConfig) (*detector, error) {

	d := &detector{
		labels: make(map[int]string),
		logger: zap.S().With("package", "detector.gocv"),
		pool:   make(chan *YOLONetwork, c.NumConcurrent),
	}

	d.config.Name = c.Name
	d.config.Type = c.Type
	d.config.Model = c.ModelFile
	d.config.Labels = make([]string, 0)

	// Load labels
	f, err := os.Open(c.LabelFile)
	if err != nil {
		return nil, fmt.Errorf("could not load label", "error", err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for x := 0; scanner.Scan(); x++ {
		fields := strings.SplitAfterN(scanner.Text(), " ", 2)
		if len(fields) == 1 {
			d.labels[x] = fields[0]
			d.config.Labels = append(d.config.Labels, fields[0])
		} else if len(fields) == 2 {
			if y, err := strconv.Atoi(strings.TrimSpace(fields[0])); err == nil {
				d.labels[y] = strings.TrimSpace(fields[1])
				d.config.Labels = append(d.config.Labels, strings.TrimSpace(fields[1]))
			}
		}
	}

	n := YOLONetwork{
		GPUDeviceIndex:           -1,
		DataConfigurationFile:    "models/coco.data",
		NetworkConfigurationFile: "models/yolov3.cfg",
		WeightsFile:              "models/yolov3.weights",
		Threshold:                .5,
	}
	if err := n.Init(); err != nil {
		return nil, fmt.Errorf("could not initialize network: %v", err)
	}

	img, err := ImageFromPath("grace_hopper.png")
	if err != nil {
		return nil, fmt.Errorf("could not load image: %v", err)
	}

	for x := 0; x < 5; x++ {
		r, err := n.Detect(img)
		if err != nil {
			return nil, fmt.Errorf("could not initialize network: %v", err)
		}
		d.logger.Infof("Detections: %v", r)
	}

	return d, nil

}

func (d *detector) Config() *odrpc.Detector {
	return &d.config
}

func (d *detector) Shutdown() {
	// d.net.Close()
}

func (d *detector) Detect(ctx context.Context, request *odrpc.DetectRequest) *odrpc.DetectResponse {

	// d.concurrent <- struct{}{}
	// defer func() {
	// 	<-d.concurrent
	// }()

	// // Read the image from the data
	// img, err := gocv.IMDecode(request.Data, gocv.IMReadColor)
	// if err != nil {
	// 	return &odrpc.DetectResponse{
	// 		Id:    request.Id,
	// 		Error: err.Error(),
	// 	}
	// }

	// blob := gocv.BlobFromImage(img, float64(1.0/255.0), image.Pt(YOLO3_FASTEST, YOLO3_FASTEST), gocv.NewScalar(0, 0, 0, 0), false, false)
	// d.net.SetInput(blob, "data")
	// defer blob.Close()

	// var outputLayers []string
	// for i := range d.net.GetUnconnectedOutLayers() {
	// 	layer := d.net.GetLayer(i)
	// 	d.logger.Infow("Processing Layer", "name", layer.GetName(), "type", layer.GetType())
	// 	layerName := layer.GetName()
	// 	if layerName != "_input" {
	// 		outputLayers = append(outputLayers, layerName)
	// 	}
	// }

	// // run a forward pass thru the network
	// prob := net.Forward("softmax2")
	// defer prob.Close()

	// // reshape the results into a 1x1000 matrix
	// probMat := prob.Reshape(1, 20)
	// defer probMat.Close()

	// // determine the most probable classification, and display it
	// _, maxVal, _, maxLoc := gocv.MinMaxLoc(probMat)
	// fmt.Printf("maxLoc: %v, maxVal: %v\n", maxLoc, maxVal)

	// detections := make([]*odrpc.Detection, 0)
	// for i := 0; i < count; i++ {
	// 	// Get the label
	// 	label, ok := d.labels[int(classes[i])]
	// 	if !ok {
	// 		d.logger.Warnw("Missing label", "index", classes[i])
	// 	}

	// 	// We have this class listed explicitly
	// 	if score, ok := request.Detect[label]; ok {
	// 		// Does it meet the score?
	// 		if scores[i]*100.0 < score {
	// 			continue
	// 		}
	// 		// We have a wildcard score
	// 	} else if score, ok := request.Detect["*"]; ok {
	// 		if scores[i]*100.0 < score {
	// 			continue
	// 		}
	// 	} else if len(request.Detect) != 0 {
	// 		// It's not listed
	// 		continue
	// 	}

	// 	detection := &odrpc.Detection{
	// 		Y1:         int32(locations[(i*4)] * float32(dy)),
	// 		X1:         int32(locations[(i*4)+1] * float32(dx)),
	// 		Y2:         int32(locations[(i*4)+2] * float32(dy)),
	// 		X2:         int32(locations[(i*4)+3] * float32(dx)),
	// 		Label:      label,
	// 		Confidence: scores[i] * 100.0,
	// 	}
	// 	// Cleanup the bounds
	// 	if detection.Y1 < 0 {
	// 		detection.Y1 = 0
	// 	}
	// 	if detection.X1 < 0 {
	// 		detection.X1 = 0
	// 	}
	// 	if detection.Y2 > dy {
	// 		detection.Y2 = dy
	// 	}
	// 	if detection.X2 > dx {
	// 		detection.X2 = dx
	// 	}
	// 	detections = append(detections, detection)

	// 	d.logger.Debugw("Detection", "id", request.Id, "label", detection.Label, "confidence", detection.Confidence, "location", fmt.Sprintf("%d,%d,%d,%d", detection.X1, detection.Y1, detection.X2, detection.Y2))
	// }

	// d.logger.Infow("Detection Complete", "id", request.Id, "duration", time.Since(start), "detections", len(detections))

	return &odrpc.DetectResponse{
		Id: request.Id,
	}
}
