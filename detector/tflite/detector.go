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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/snowzach/doods/conf"
	"github.com/snowzach/doods/detector/dconfig"
	"github.com/snowzach/doods/odrpc"

	"github.com/snowzach/doods/detector/tflite/go-tflite"
	"github.com/snowzach/doods/detector/tflite/go-tflite/delegates/edgetpu"
)

const (
	OutputFormat_4_TFLite_Detection_PostProcess = iota
	OutputFormat_1_scores
)

type detector struct {
	config odrpc.Detector
	logger *zap.SugaredLogger

	labels       map[int]string
	model        *tflite.Model
	inputType    tflite.TensorType
	outputFormat int
	pool         chan *tflInterpreter

	devices    []edgetpu.Device
	numThreads int
	hwAccel    bool
	timeout    time.Duration
}

type tflInterpreter struct {
	device *edgetpu.Device
	*tflite.Interpreter
}

func New(c *dconfig.DetectorConfig) (*detector, error) {

	d := &detector{
		labels:     make(map[int]string),
		logger:     zap.S().With("package", "detector.tflite", "name", c.Name),
		pool:       make(chan *tflInterpreter, c.NumConcurrent),
		numThreads: c.NumThreads,
		hwAccel:    c.HWAccel,
		timeout:    c.Timeout,
	}

	d.config.Name = c.Name
	d.config.Type = c.Type
	d.config.Model = c.ModelFile
	d.config.Labels = make([]string, 0)

	// Create the model
	d.model = tflite.NewModelFromFile(d.config.Model)
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
	if d.hwAccel {

		// Get the list of devices
		d.devices, err = edgetpu.DeviceList()
		if err != nil {
			return nil, fmt.Errorf("Could not fetch edgetpu device list: %v", err)
		}
		if len(d.devices) == 0 {
			return nil, fmt.Errorf("no edgetpu devices detected")
		}
		c.NumConcurrent = len(d.devices)
		d.config.Type = "tflite-edgetpu"

		// Enforce a timeout for edgetpu devices if not set
		if d.timeout == 0 {
			d.timeout = 30 * time.Second
		}

	}

	// Create the pool of interpreters
	var interpreter *tflInterpreter
	for x := 0; x < c.NumConcurrent; x++ {

		interpreter = new(tflInterpreter)

		// Get a device if there is one
		if d.hwAccel && len(d.devices) > x {
			interpreter.device = &d.devices[x]
		}

		interpreter.Interpreter, err = d.newInterpreter(interpreter.device)
		if err != nil {
			return nil, err
		}

		d.pool <- interpreter
	}

	// Get the settings from the input tensor
	if inputCount := interpreter.GetInputTensorCount(); inputCount != 1 {
		return nil, fmt.Errorf("unsupported input tensor count: %d", inputCount)
	}
	input := interpreter.GetInputTensor(0)
	if input.Name() != "normalized_input_image_tensor" && input.Name() != "image" {
		return nil, fmt.Errorf("unsupported input tensor name: %s", input.Name())
	}
	d.config.Height = int32(input.Dim(1))
	d.config.Width = int32(input.Dim(2))
	d.config.Channels = int32(input.Dim(3))
	d.inputType = input.Type()
	if d.inputType != tflite.UInt8 {
		return nil, fmt.Errorf("unsupported tensor input type: %s", d.inputType)
	}

	// Dump output tensor information
	count := interpreter.GetOutputTensorCount()
	for x := 0; x < count; x++ {
		tensor := interpreter.GetOutputTensor(x)
		numDims := tensor.NumDims()
		d.logger.Debugw("Tensor Output", "n", x, "name", tensor.Name(), "type", tensor.Type(), "num_dims", numDims, "byte_size", tensor.ByteSize())
		if numDims > 1 {
			for y := 0; y < numDims; y++ {
				d.logger.Debugw("Tensor Dim", "n", x, "dim", y, "dim_size", tensor.Dim(x))
			}
		}
	}

	if count == 4 && interpreter.GetOutputTensor(0).Name() == "TFLite_Detection_PostProcess" {
		d.outputFormat = OutputFormat_4_TFLite_Detection_PostProcess
	} else if count == 1 && interpreter.GetOutputTensor(0).Name() == "scores" {
		d.outputFormat = OutputFormat_1_scores
		// Check the output types
		tensor := interpreter.GetOutputTensor(0)
		if tensor.Type() != tflite.UInt8 {
			return nil, fmt.Errorf("unsupported tensor output type: %s", tensor.Type())
		}
		// Ensure the length of the labels match the detection size
		for x := int(tensor.ByteSize()) - len(d.labels); x > 0; x-- {
			d.labels[x] = "unknown"
		}
	} else {
		return nil, fmt.Errorf("unsupported output tensor count: %d", count)
	}

	return d, nil
}

func (d *detector) newInterpreter(device *edgetpu.Device) (*tflite.Interpreter, error) {
	// Options
	options := tflite.NewInterpreterOptions()
	options.SetNumThread(d.numThreads)
	options.SetErrorReporter(func(msg string, user_data interface{}) {
		d.logger.Warnw("Error", "message", msg, "user_data", user_data)
	}, nil)

	// Use edgetpu
	if device != nil {
		etpuInstance := edgetpu.New(*device)
		if etpuInstance == nil {
			return nil, fmt.Errorf("could not initialize edgetpu %s", device.Path)
		}
		options.AddDelegate(etpuInstance)
	}

	interpreter := tflite.NewInterpreter(d.model, options)
	if interpreter == nil {
		return nil, fmt.Errorf("Could not create interpreter")
	}

	// Allocate
	status := interpreter.AllocateTensors()
	if status != tflite.OK {
		return nil, fmt.Errorf("interpreter allocate failed")
	}

	return interpreter, nil
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

func (d *detector) Detect(ctx context.Context, request *odrpc.DetectRequest) (*odrpc.DetectResponse, error) {

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
			return nil, status.Errorf(codes.InvalidArgument, "could not decode image: %v", err)
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
	conf.Stop.Add(1) // Wait until detection complete before stopping
	defer func() {
		d.pool <- interpreter
		conf.Stop.Done()
	}()

	// Build the tensor input
	input := interpreter.GetInputTensor(0)
	input.CopyFromBuffer(data)

	start := time.Now()

	// Perform the detection
	var invokeStatus tflite.Status
	complete := make(chan struct{})
	go func() {
		invokeStatus = interpreter.Invoke()
		close(complete)
	}()

	// Wait for complete or timeout if there is one set
	if d.timeout > 0 {
		select {
		case <-complete:
			// We're done
		case <-time.After(d.timeout):
			// The detector is hung, it needs to be reinitialized
			d.logger.Errorw("Detector timeout", zap.Any("device", interpreter.device))
			conf.Stop.Stop() // Exit after all threads complete
			return nil, status.Errorf(codes.Internal, "detect failed")
		}
	}
	<-complete // Complete no timeout

	// Capture Errors
	if invokeStatus != tflite.OK {
		d.logger.Errorw("Detector error", "id", request.Id, "status", invokeStatus, zap.Any("device", interpreter.device))
		return &odrpc.DetectResponse{
			Id:    request.Id,
			Error: "detector error",
		}, nil
	}

	detections := make([]*odrpc.Detection, 0)

	switch d.outputFormat {
	case OutputFormat_4_TFLite_Detection_PostProcess:
		// Parse results
		var countResult float32
		interpreter.GetOutputTensor(3).CopyToBuffer(&countResult)
		count := int(countResult)

		// Check for a sane count value
		if count < 0 || count > 100 {
			d.logger.Errorw("Detector invalid results", "id", request.Id, "count", count, zap.Any("device", interpreter.device))
			return &odrpc.DetectResponse{
				Id:    request.Id,
				Error: "detector invalid result",
			}, nil
		}

		locations := make([]float32, count*4, count*4)
		classes := make([]float32, count, count)
		scores := make([]float32, count, count)

		if count > 0 {
			interpreter.GetOutputTensor(0).CopyToBuffer(&locations[0])
			interpreter.GetOutputTensor(1).CopyToBuffer(&classes[0])
			interpreter.GetOutputTensor(2).CopyToBuffer(&scores[0])
		}

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

			d.logger.Debugw("Detection", "id", request.Id, "label", detection.Label, "confidence", detection.Confidence, "location", fmt.Sprintf("%f,%f,%f,%f", detection.Top, detection.Left, detection.Bottom, detection.Right))
		}

	case OutputFormat_1_scores:
		scores := make([]uint8, len(d.labels), len(d.labels))
		interpreter.GetOutputTensor(0).CopyToBuffer(scores)

		for i := range scores {
			// Get the label
			label, ok := d.labels[i]
			if !ok {
				d.logger.Warnw("Missing label", "index", i)
			}

			labelScore := 100.0 * (float32(scores[i]) / 255.0)

			// We have this class listed explicitly
			if matchScore, ok := request.Detect[label]; ok {
				// Does it meet the score?
				if labelScore < matchScore {
					continue
				}
				// We have a wildcard score
			} else if matchScore, ok := request.Detect["*"]; ok {
				if labelScore < matchScore {
					continue
				}
			} else if len(request.Detect) != 0 {
				// It's not listed
				continue
			}

			detection := &odrpc.Detection{
				Top:        0.0,
				Left:       0.0,
				Bottom:     1.0,
				Right:      1.0,
				Label:      label,
				Confidence: labelScore,
			}
			detections = append(detections, detection)
			d.logger.Debugw("Detection", "id", request.Id, "label", detection.Label, "confidence", detection.Confidence, "location", fmt.Sprintf("%f,%f,%f,%f", detection.Top, detection.Left, detection.Bottom, detection.Right))
		}
	}

	d.logger.Infow("Detection Complete", "id", request.Id, "duration", time.Since(start), "detections", len(detections), zap.Any("device", interpreter.device))

	return &odrpc.DetectResponse{
		Id:         request.Id,
		Detections: detections,
	}, nil
}
