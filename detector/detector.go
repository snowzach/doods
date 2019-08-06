package detector

import (
	"context"
	"fmt"
	"sync"

	// We will support these formats
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "github.com/lmittmann/ppm"
	_ "golang.org/x/image/bmp"

	emptypb "github.com/golang/protobuf/ptypes/empty"
	config "github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/snowzach/doods/detector/dconfig"
	"github.com/snowzach/doods/detector/tflite"
	"github.com/snowzach/doods/detector/tensorflow"
	"github.com/snowzach/doods/detector/darknet"
	"github.com/snowzach/doods/odrpc"
)

// Detector is the interface to object detectors
type Detector interface {
	Config() *odrpc.Detector
	Detect(ctx context.Context, request *odrpc.DetectRequest) *odrpc.DetectResponse
	Shutdown()
}

// Mux handles and routes requests to the configured detectors
type Mux struct {
	detectors map[string]Detector
	authKey   string
	logger    *zap.SugaredLogger
}

// Create a new mux
func New() *Mux {

	m := &Mux{
		detectors: make(map[string]Detector),
		authKey:   config.GetString("doods.auth_key"),
		logger:    zap.S().With("package", "detector"),
	}

	// Get the detectors config
	var detectorConfig []*dconfig.DetectorConfig
	config.UnmarshalKey("doods.detectors", &detectorConfig)

	// Create the detectors
	for _, c := range detectorConfig {
		var d Detector
		var err error

		m.logger.Debugw("Configuring detector", "config", c)

		switch c.Type {
		case "tflite":
			d, err = tflite.New(c)
		case "tensorflow":
			d, err = tensorflow.New(c)
		case "darknet":
			d, err = darknet.New(c)
		default:
			m.logger.Errorw("Unknown detector", "name", c.Name, "type", c.Type)
			continue
		}

		if err != nil {
			m.logger.Errorf("Could not initialize detector %s: %v", c.Name, err)
			continue
		}

		dc := d.Config()
		m.logger.Infow("Configured Detector", "name", dc.Name, "type", dc.Type, "model", dc.Model, "labels", len(dc.Labels), "width", dc.Width, "height", dc.Height)
		m.detectors[c.Name] = d
	}

	if len(m.detectors) == 0 {
		m.logger.Fatalf("No detectors configured")
	}

	return m

}

// GetDetectors returns the configured detectors
func (m *Mux) GetDetectors(ctx context.Context, _ *emptypb.Empty) (*odrpc.GetDetectorsResponse, error) {
	detectors := make([]*odrpc.Detector, 0)
	for _, d := range m.detectors {
		detectors = append(detectors, d.Config())
	}
	return &odrpc.GetDetectorsResponse{
		Detectors: detectors,
	}, nil
}

// Shutdown deallocates/shuts down any detectors
func (m *Mux) Shutdown() {
	for _, d := range m.detectors {
		d.Shutdown()
	}
}

// Run a detection
func (m *Mux) Detect(ctx context.Context, request *odrpc.DetectRequest) (*odrpc.DetectResponse, error) {

	if request.DetectorName == "" {
		request.DetectorName = "default"
	}

	detector, ok := m.detectors[request.DetectorName]
	if !ok {
		return &odrpc.DetectResponse{
			Id:    request.Id,
			Error: fmt.Sprintf("unknown detector %s", request.DetectorName),
		}, nil
	}

	return detector.Detect(ctx, request), nil

}

// Handle a stream of detections
func (m *Mux) DetectStream(stream odrpc.Odrpc_DetectStreamServer) error {

	ctx := stream.Context()
	var send sync.Mutex

	for ctx.Err() == nil {

		request, err := stream.Recv()
		if err != nil {
			return nil
		}

		m.logger.Info("Request")

		go func() {
			response, _ := m.Detect(ctx, request)
			send.Lock()
			stream.Send(response)
			send.Unlock()
		}()

	}

	return nil

}
