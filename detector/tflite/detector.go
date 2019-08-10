package tflite

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"image"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nfnt/resize"
	"go.uber.org/zap"

	"github.com/snowzach/doods/detector/dconfig"
	"github.com/snowzach/doods/odrpc"
)

type detector struct {
	config odrpc.Detector
	logger *zap.SugaredLogger

	labels    map[int]string
	model     *Model
	inputType TensorType
	pool      chan *Interpreter
}

func New(c *dconfig.DetectorConfig) (*detector, error) {

	d := &detector{
		labels: make(map[int]string),
		logger: zap.S().With("package", "detector.tflite"),
		pool:   make(chan *Interpreter, c.NumConcurrent),
	}

	d.config.Name = c.Name
	d.config.Type = c.Type
	d.config.Model = c.ModelFile
	d.config.Labels = make([]string, 0)

	// Create the model
	d.model = NewModelFromFile(d.config.Model)
	if d.model == nil {
		return nil, fmt.Errorf("could not load model %s", d.config.Model)
	}

	// Load labels
	f, err := os.Open(c.LabelFile)
	if err != nil {
		return nil, fmt.Errorf("could not load label", "error", err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for x := 1; scanner.Scan(); x++ {
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

	// If we are using edgetpu, make sure we have one
	if c.HWAccel {
		if !HasEdgeTPU() {
			return nil, fmt.Errorf("no edgetpu detected")
		}
		d.config.Type = "tflite-edgetpu"
		d.logger.Warn("Warnings about Failed to retrieve TPU context are normal")
	}

	var interpreter *Interpreter

	for x := 0; x < c.NumConcurrent; x++ {
		// Options
		options := NewInterpreterOptions()
		options.SetNumThread(c.NumThreads)
		options.SetErrorReporter(func(msg string, user_data interface{}) {
			d.logger.Warnw("Error", "message", msg, "user_data", user_data)
		}, nil)
		defer options.Delete()

		// Use edgetpu
		if c.HWAccel {
			options.ExpAddCustomOp(EdgeTPUCustomOp, Register_EdgeTPU(), 1, 1)
			interpreter = NewInterpreter(d.model, options)
			EdgeTPUSetup(interpreter, d.model)
		} else {
			interpreter = NewInterpreter(d.model, options)
		}

		if interpreter == nil {
			return nil, fmt.Errorf("Could not create interpreter")
		}

		// Allocate
		status := interpreter.AllocateTensors()
		if status != OK {
			return nil, fmt.Errorf("interpreter allocate failed")
		}

		d.pool <- interpreter

	}

	// Get the settings
	input := interpreter.GetInputTensor(0)
	d.config.Height = int32(input.Dim(1))
	d.config.Width = int32(input.Dim(2))
	d.config.Channels = int32(input.Dim(3))
	d.inputType = input.Type()
	if d.inputType != UInt8 {
		return nil, fmt.Errorf("unsupported tensor input type: %s", d.inputType)
	}

	return d, nil
}

func (d *detector) Config() *odrpc.Detector {
	return &d.config
}

func (d *detector) Shutdown() {
	close(d.pool)
	for {
		interpreter := <-d.pool
		if interpreter == nil {
			break
		}
		interpreter.Delete()
	}
}

func (d *detector) Detect(ctx context.Context, request *odrpc.DetectRequest) *odrpc.DetectResponse {

	var data []byte
	var dx, dy int32

	// If this is ppm data, move it right to tensorflow
	if ppmInfo := FindPPMData(request.Data); ppmInfo != nil && int32(ppmInfo.Width) == d.config.Width && int32(ppmInfo.Height) == d.config.Height {
		dx, dy = d.config.Width, d.config.Height
		// Dump data right to data input
		data = request.Data[ppmInfo.Offset:]
	} else {
		// Decode the image
		img, format, err := image.Decode(bytes.NewReader(request.Data))
		if err != nil {
			return &odrpc.DetectResponse{
				Id:    request.Id,
				Error: fmt.Sprintf("Could not decode: %v", err),
			}
		}

		// Resize it if necessary
		bounds := img.Bounds()
		dx = int32(bounds.Dx())
		dy = int32(bounds.Dy())
		d.logger.Debugw("Got Image", "id", request.Id, "format", format, "width", dx, "height", dy)
		if dx != int32(d.config.Width) || dy != int32(d.config.Height) {
			d.logger.Debugw("Resizing Image", "id", request.Id, "format", format, "width", d.config.Width, "height", d.config.Height)
			img = resize.Resize(uint(d.config.Width), uint(d.config.Height), img, resize.NearestNeighbor)
		}

		data = make([]byte, int(d.config.Width*d.config.Height*d.config.Channels))
		for y := int32(0); y < d.config.Height; y++ {
			for x := int32(0); x < d.config.Width; x++ {
				col := img.At(int(x), int(y))
				r, g, b, _ := col.RGBA()
				data[(y*d.config.Width+x)*3+0] = byte(float64(r) / 255.0)
				data[(y*d.config.Width+x)*3+1] = byte(float64(g) / 255.0)
				data[(y*d.config.Width+x)*3+2] = byte(float64(b) / 255.0)
			}
		}
	}

	// Get an interpreter from the pool
	interpreter := <-d.pool
	defer func() {
		d.pool <- interpreter
	}()

	// Build the tensor input
	input := interpreter.GetInputTensor(0)
	input.CopyFromBuffer(data)

	start := time.Now()

	invokeStatus := interpreter.Invoke()
	if invokeStatus != OK {
		return &odrpc.DetectResponse{
			Id:    request.Id,
			Error: fmt.Sprintf("detect failed"),
		}
	}

	countResult := make([]float32, 1, 1)
	interpreter.GetOutputTensor(3).CopyToBuffer(&countResult[0])
	count := int(countResult[0])

	locations := make([]float32, count*4, count*4)
	interpreter.GetOutputTensor(0).CopyToBuffer(&locations[0])

	classes := make([]float32, count, count)
	interpreter.GetOutputTensor(1).CopyToBuffer(&classes[0])

	scores := make([]float32, count, count)
	interpreter.GetOutputTensor(2).CopyToBuffer(&scores[0])

	detections := make([]*odrpc.Detection, 0)
	for i := 0; i < count; i++ {
		// Get the label
		label, ok := d.labels[int(classes[i])]
		if !ok {
			d.logger.Warnw("Missing label", "index", classes[i])
		}

		// We have this class listed explicitly
		if score, ok := request.Detect[label]; ok {
			// Does it meet the score?
			if scores[i]*100.0 < score {
				continue
			}
			// We have a wildcard score
		} else if score, ok := request.Detect["*"]; ok {
			if scores[i]*100.0 < score {
				continue
			}
		} else if len(request.Detect) != 0 {
			// It's not listed
			continue
		}

		detection := &odrpc.Detection{
			Top:        locations[(i * 4)],
			Left:       locations[(i*4)+1],
			Bottom:     locations[(i*4)+2],
			Right:      locations[(i*4)+3],
			Label:      label,
			Confidence: scores[i] * 100.0,
		}
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
		detections = append(detections, detection)

		d.logger.Debugw("Detection", "id", request.Id, "label", detection.Label, "confidence", detection.Confidence, "location", fmt.Sprintf("%d,%d,%d,%d", detection.Top, detection.Left, detection.Bottom, detection.Right))
	}

	d.logger.Infow("Detection Complete", "id", request.Id, "duration", time.Since(start), "detections", len(detections))

	return &odrpc.DetectResponse{
		Id:         request.Id,
		Detections: detections,
	}
}
